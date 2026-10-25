# Step 1 — Task board: one service, one database

**Epic:** DS-E1 — Order Service + Postgres in Docker
**Goal:** One command (`docker compose up`) starts Postgres and the Order Service.
I can create an order and read it over HTTP. I know what happens when the database stops.

| Key | Title | Type | Who | Status |
|---|---|---|---|---|
| DS-1 | Install WSL2 + Docker Desktop | Task | 👤 Me | ✅ Done |
| DS-2 | Order Service Go code | Story | 🤖 Claude | ✅ Done (I must review it) |
| DS-3 | SQL script for the `orders` table | Task | 👤 Me | 🔲 To Do |
| DS-4 | Multi-stage Dockerfile | Task | 👤 Me | 🔲 To Do |
| DS-5 | `docker-compose.yml` | Task | 👤 Me | 🔲 To Do |
| DS-6 | Experiment: stop the database | Spike | 👤 Me | 🔲 To Do |

**Order:** DS-2 review → DS-3 → DS-4 → DS-5 → DS-6. (DS-5 needs DS-3 and DS-4.)

> **How to use the hints:** First try without hints.
> Stuck for 15+ minutes? Open **Hint 1**. Still stuck? Open **Hint 2**.
> A little struggle is good — this is how you learn.

---

## DS-1 — Install WSL2 + Docker Desktop

**Type:** Task · **Who:** 👤 Me · **Status:** ✅ Done

Checked: WSL 2, Docker 29.7.2, Compose v5.3.1, Go 1.27.1. The `hello-world` container runs.

**Remember:** Docker Desktop must be **open** before you use any `docker` command.

---

## DS-2 — Order Service Go code

**Type:** Story · **Who:** 🤖 Claude writes, 👤 I review · **Status:** ✅ Done (I must review it)

**Story**
As a client, I want to create an order and get it by its id over HTTP.
The order must be saved in Postgres, so it is not lost.

**Files**
```
order-service/
├── go.mod / go.sum
├── main.go                     # starts everything: config → pool → router → HTTP server
└── internal/
    ├── config/config.go        # reads PORT (default 8080) and DATABASE_URL (required)
    └── order/
        ├── order.go            # the Order struct and status values
        ├── repository.go       # all SQL is here (uses pgxpool)
        └── handler.go          # HTTP: checks input, returns status codes, DB timeout
```

**API**

| Method | Path | Body | OK | Errors |
|---|---|---|---|---|
| POST | `/orders` | `{"user_id": "u1", "amount": 1999}` | `201` + order | `400` bad input, `500` DB error |
| GET | `/orders/:id` | — | `200` + order | `400` id is not a UUID, `404` not found, `500` |
| GET | `/health` | — | `200 {"status":"ok","db":"up"}` | `503` DB not reachable |

`amount` is in **cents**: `1999` = 19.99.

📖 **Read first:** [docs/database.md](../docs/database.md) — how we connect to Postgres (no ORM, pool, login, security).

**Review checklist** — I can answer these out loud, without looking:
- [ ] Follow `POST /orders` from the HTTP request to the SQL `INSERT` and back.
- [ ] Why does `main.go` check the DB (`Ping`) at start?
- [ ] Why does every query have `context.WithTimeout`? What happens without it?
- [ ] Why is `amount` an `int64` in cents and not a `float64`?
- [ ] For a `500` error, why do we hide the real error from the client but write it to the log?
- [ ] What does `pgxpool` give us? Why not one single connection?
- [ ] What happens on `docker stop`? Why is a "graceful shutdown" important?

---

## DS-3 — SQL script for the `orders` table

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do

**Description**
The service needs an `orders` table in Postgres. Write a simple SQL script that creates it.
Postgres will run this script by itself on the first start (you connect it in DS-5).

**File:** `order-service/db/init.sql`

**Acceptance criteria**
- [ ] The table looks **exactly** like this (the Go code needs these names and types):

  | Column | Rules |
  |---|---|
  | `id` | UUID, primary key, Postgres makes it |
  | `user_id` | text, cannot be empty (not null) |
  | `amount` | 64-bit whole number (cents), not null, must be bigger than 0 |
  | `status` | text, not null |
  | `created_at` | time **with time zone**, not null, default = now |

- [ ] If the script runs two times, it does not fail.
- [ ] I can answer: Why does the DB make the `id`, and not Go? Why `TIMESTAMPTZ` and not `TIMESTAMP`?
  Why check `amount > 0` in the DB, when Go already checks it?

<details>
<summary>Hint 1</summary>

- "Runs two times without failing" → search `CREATE TABLE IF NOT EXISTS`.
- Postgres has a built-in function that makes a random UUID → search `gen_random_uuid`.
- A 64-bit whole number in Postgres is `BIGINT`.
- A rule like "must be bigger than 0" is a `CHECK (...)` constraint.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```sql
CREATE TABLE IF NOT EXISTS orders (
    id          ____ PRIMARY KEY DEFAULT ____,
    user_id     ____ NOT NULL,
    amount      ____ NOT NULL CHECK (____),
    status      ____ NOT NULL,
    created_at  ____ NOT NULL DEFAULT ____
);
```
</details>

---

## DS-4 — Multi-stage Dockerfile

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do

**Description**
Put the Order Service into a small and safe Docker image.
Use a **multi-stage build**, which means two stages:

1. **Build stage:** has the full Go tools. It compiles the code into one binary file.
2. **Run stage:** a very small image. It has **only** the binary — no Go tools, no source code.

**Files:** `order-service/Dockerfile` and `order-service/.dockerignore`

**Acceptance criteria**
- [ ] The build stage uses an official `golang` image with the same version as `go.mod` (1.27). The stage has a name (for example `AS build`).
- [ ] First copy `go.mod` and `go.sum`, run `go mod download`, and **after that** copy the rest of the code.
- [ ] The binary is **static** (it needs no C libraries).
- [ ] The run stage uses a very small base image and has **only** the binary.
- [ ] The container does **not** run as `root`.
- [ ] `EXPOSE 8080`, and the binary is the `ENTRYPOINT`.
- [ ] `.dockerignore` keeps unneeded files out of the build (`.git`, binaries, notes, editor files).
- [ ] `docker build -t order-service ./order-service` works.
- [ ] I wrote down two sizes: my final image, and the `golang` base image.
- [ ] I can answer: Why is the final image so much smaller? Why copy `go.mod` first (what is layer caching)? Why not run as root?

<details>
<summary>Hint 1</summary>

- Multi-stage = more than one `FROM` line. A later stage can take files from an earlier stage with `COPY --from=<stage-name>`.
- Static Go binary: set `CGO_ENABLED=0` and `GOOS=linux` when you run `go build`.
- Small run images: `gcr.io/distroless/static-debian12` (very small, no shell) or `alpine` (small, has a shell — good for debugging).
- Distroless has a version that is already non-root. Look at its tags.
- Bonus: what does `go build -ldflags="-s -w"` do?
- See image sizes with `docker images`.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```dockerfile
# ---- build stage ----
FROM golang:____ AS build
WORKDIR /src
COPY ____ ____ ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-service .

# ---- run stage ----
FROM ____
COPY --from=____ /out/order-service /order-service
USER ____
EXPOSE ____
ENTRYPOINT ["____"]
```
</details>

---

## DS-5 — `docker-compose.yml`

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-3, DS-4

**Description**
One command, `docker compose up --build`, starts Postgres and the Order Service.
The table is created by itself. The service finds Postgres by its **service name**.

📖 Read [docs/database.md](../docs/database.md) sections 3 and 4 first (connection URL and login).

**File:** `docker-compose.yml` (in the project root folder)

**Acceptance criteria**
- [ ] A `postgres` service: official image (for example `postgres:17`), with user, password and database name set by env vars.
- [ ] Your `init.sql` (DS-3) is mounted, so Postgres runs it on the first start.
- [ ] Postgres data is in a **named volume**, so it is not lost after `docker compose down`.
- [ ] Postgres has a **healthcheck**.
- [ ] An `order-service` service: builds from `./order-service`, publishes port `8080`, and gets `DATABASE_URL` as an env var.
- [ ] `order-service` waits until Postgres is **healthy** (ready to work), not only **started**.
- [ ] `curl.exe http://localhost:8080/health` returns `200`.
- [ ] `POST /orders`, then `GET /orders/:id` returns the same order.
- [ ] I can answer: Why is the DB host not `localhost`? Does the init script run on every start, or only once?
  What is the difference between "started" and "healthy"?

<details>
<summary>Hint 1</summary>

- Postgres image env vars: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`.
- The image runs every `.sql` file in `/docker-entrypoint-initdb.d/` — but **only if the data folder is empty** (first start).
- Postgres has a tool to check if it is ready: `pg_isready`.
- Wait for healthy: `depends_on` with `condition: service_healthy`.
- In Compose, each service can reach the others by **service name**.
- URL format: `postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable`
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```yaml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_USER: ____
      POSTGRES_PASSWORD: ____
      POSTGRES_DB: ____
    volumes:
      - ____:/var/lib/postgresql/data
      - ./order-service/db/init.sql:____
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ____ -d ____"]
      interval: 5s
      timeout: 3s
      retries: 10

  order-service:
    build: ____
    ports:
      - "____:____"
    environment:
      DATABASE_URL: postgres://____:____@____:5432/____?sslmode=disable
    depends_on:
      postgres:
        condition: ____

volumes:
  ____:
```

Test commands (in PowerShell use `curl.exe`, not `curl`):

```powershell
curl.exe -i http://localhost:8080/health
curl.exe -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{\"user_id\":\"u1\",\"amount\":1999}'
curl.exe -i http://localhost:8080/orders/<id-from-above>
```
</details>

---

## DS-6 — Experiment: stop the database

**Type:** Spike · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-5

**Description**
Break the database on purpose and watch what the service does.
Write what you see in `notes/step1.md`.

**Acceptance criteria**
- [ ] Everything is running. Stop only Postgres: `docker compose stop postgres`.
- [ ] Write down: what do `GET /health`, `POST /orders` and `GET /orders/:id` return? (status code + body)
- [ ] Write down: what do the service logs say? (`docker compose logs -f order-service`)
- [ ] How long does one request take while the DB is down? Why this long? (Look for a timeout in `handler.go`.)
- [ ] Start Postgres again: `docker compose start postgres`. Do **not** restart the service. Does it work again by itself? Why?
- [ ] Bonus: stop Postgres, then restart the **Order Service**. What happens? Why is it different?
- [ ] Bonus: `docker compose down`, then `up`. Is the data still there? Now `down -v`, then `up`. What changed?

<details>
<summary>Hint 1</summary>

- Measure the time of a request: `curl.exe -w "%{time_total}\n" ...`
- `pgxpool` opens connections **when they are needed**. What does this mean when Postgres comes back?
- At start, `main.go` checks the DB and stops if it fails. Compare this with "the DB stops while the service is running".
</details>

---

## Definition of Done (Step 1)

- [ ] DS-1 … DS-6 are done. Claude checked DS-3, DS-4 and DS-5 in the terminal.
- [ ] `notes/step1.md` has my experiment results.
- [ ] I can explain without looking: request → handler → DB → response; what a connection pool is and why we need it;
  why multi-stage builds make images smaller; how containers find each other by name in Compose.
- [ ] Committed and pushed to git.
