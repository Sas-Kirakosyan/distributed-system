# Step 1 — Task board: one service, one database

**Epic:** DS-E1 — Order Service + Postgres in Docker
**Goal:** One command (`docker compose up`) starts Postgres and the Order Service.
I can create an order and read it over HTTP. I know what happens when the database stops.

| Key  | Title                             | Type  | Who       | Status                     |
| ---- | --------------------------------- | ----- | --------- | -------------------------- |
| DS-1 | Install WSL2 + Docker Desktop     | Task  | 👤 Me     | ✅ Done                    |
| DS-2 | Order Service Go code             | Story | 🤖 Claude | ✅ Done                    |
| DS-3 | SQL script for the `orders` table | Task  | 👤 Me     | 🔲 To Do                   |
| DS-4 | Multi-stage Dockerfile            | Task  | 👤 Me     | 🔲 To Do                   |
| DS-5 | `docker-compose.yml`              | Task  | 👤 Me     | 🔲 To Do                   |
| DS-6 | Experiment: stop the database     | Spike | 👤 Me     | 🔲 To Do                   |
| DS-7 | Install Ubuntu in WSL             | Task  | 👤 Me     | ✅ Done                    |
| DS-8 | Linux: files and search           | Task  | 👤 Me     | ✅ Done                    |
| DS-9 | Experiment: container vs volume   | Spike | 👤 Me     | 🔲 To Do                   |
| DS-10 | Linux: env, processes and ports  | Task  | 👤 Me     | 🔲 To Do                   |
| DS-11 | Linux: permissions and pipes     | Task  | 👤 Me     | 🔲 To Do                   |
| DS-12 | Linux: look inside a container   | Task  | 👤 Me     | 🔲 To Do                   |
| DS-13 | Bash script: `smoke-test.sh`     | Task  | 👤 Me     | 🔲 To Do                   |

**Order:** DS-7 → DS-8 → DS-2 review → DS-3 → DS-4 → DS-5 → DS-6 → DS-9 → DS-10 → DS-11 → DS-12 → DS-13.
(DS-5 needs DS-3 and DS-4. DS-6, DS-9, DS-11, DS-12 and DS-13 need DS-5.)

> **Terminal:** Run every command in the **Ubuntu terminal**, unless a ticket says something else.
> Go to the project folder with `cd /mnt/c/Users/saski/OneDrive/Desktop/destrubuted-system`.

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

**Type:** Story · **Who:** 🤖 Claude writes, 👤 I review · **Status:** ✅ Done

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

| Method | Path          | Body                                | OK                              | Errors                                         |
| ------ | ------------- | ----------------------------------- | ------------------------------- | ---------------------------------------------- |
| POST   | `/orders`     | `{"user_id": "u1", "amount": 1999}` | `201` + order                   | `400` bad input, `500` DB error                |
| GET    | `/orders/:id` | —                                   | `200` + order                   | `400` id is not a UUID, `404` not found, `500` |
| GET    | `/health`     | —                                   | `200 {"status":"ok","db":"up"}` | `503` DB not reachable                         |

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

### What is `init.sql`?

- It is a **normal text file with SQL commands**. Nothing special inside.
- The name `init.sql` is only a habit. Postgres does not need this exact name.
- The **official Postgres Docker image** has a special folder: `/docker-entrypoint-initdb.d/`.
- When the container starts **for the first time** (the data folder is empty), Postgres:
  1. creates the user and the database (from `POSTGRES_USER`, `POSTGRES_DB`),
  2. runs **every** `.sql` file in that folder, in name order (`01-...`, `02-...`),
  3. then starts normally.
- On the **next** starts, the data already exists, so the scripts **do not run again**.
- So `init.sql` is "the first setup of the database". In DS-5 you put the file into that folder.

```
first start:   empty data folder → create user + DB → run init.sql → ready
next starts:   data folder has data → skip init.sql → ready
```

### Steps to do

1. Create the folder `order-service/db/`.
2. Create the file `order-service/db/init.sql`.
3. Write one `CREATE TABLE` command. Go row by row through the table below: one column = one line.
4. **Test the file alone** (no compose yet). Run these in the Ubuntu terminal from the project root folder:

   ```bash
   # start a temporary Postgres (wait ~5 seconds after this)
   docker run -d --name pg-test -e POSTGRES_PASSWORD=test postgres:17

   # run your script
   cat order-service/db/init.sql | docker exec -i pg-test psql -U postgres

   # run it a second time — it must NOT fail
   cat order-service/db/init.sql | docker exec -i pg-test psql -U postgres

   # look at the table
   docker exec -it pg-test psql -U postgres -c "\d orders"
   ```

5. **Test the rules.** Try to insert a bad row. It must fail:
   ```bash
   docker exec -it pg-test psql -U postgres -c "INSERT INTO orders (user_id, amount, status) VALUES ('u1', -5, 'PENDING');"
   ```
   Then insert a good row (`amount` = 1999) and check that `id` and `created_at` are filled by themselves:
   ```bash
   docker exec -it pg-test psql -U postgres -c "SELECT * FROM orders;"
   ```
6. Clean up: `docker rm -f pg-test`
7. Push your work or paste it to Claude → Claude reviews it.

**Acceptance criteria**

- [ ] The table looks **exactly** like this (the Go code needs these names and types):

  | Column       | Rules                                                        |
  | ------------ | ------------------------------------------------------------ |
  | `id`         | UUID, primary key, Postgres makes it                         |
  | `user_id`    | text, cannot be empty (not null)                             |
  | `amount`     | 64-bit whole number (cents), not null, must be bigger than 0 |
  | `status`     | text, not null                                               |
  | `created_at` | time **with time zone**, not null, default = now             |

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

### Words you need

| Word                | Meaning                                                                                                           |
| ------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Dockerfile**      | A recipe. Each line is one step to build an image.                                                                |
| **Image**           | The result of the recipe. A packed file system + the command to start.                                            |
| **Container**       | A running image. You can start many containers from one image.                                                    |
| **Layer**           | Each step in the Dockerfile makes one layer. Docker **saves** (caches) layers and reuses them if nothing changed. |
| **Build context**   | The folder you send to Docker when you build (here: `./order-service`).                                           |
| **`.dockerignore`** | Like `.gitignore`, but for the build context. Files in it are not sent to Docker.                                 |

**Why two stages?** The `golang` image is big (~800 MB). It has the compiler, tools and source code.
To _run_ the program, you only need the one binary file (~15 MB). So: build in a big image, copy only the binary into a tiny image, and throw away the big one.

### Steps to do

1. Create the file `order-service/Dockerfile` (no file extension).
2. Write the **build stage**, in this order:
   1. `FROM` the golang image, and give the stage a name.
   2. `WORKDIR` — choose a folder inside the image.
   3. `COPY` only `go.mod` and `go.sum`.
   4. `RUN go mod download`.
   5. `COPY` the rest of the code.
   6. `RUN go build ...` with `CGO_ENABLED=0 GOOS=linux`.
3. Write the **run stage**:
   1. A second `FROM` with a tiny image.
   2. `COPY --from=<build stage>` — only the binary.
   3. `USER` — a non-root user.
   4. `EXPOSE 8080`.
   5. `ENTRYPOINT` — start the binary.
4. Create `order-service/.dockerignore`. Think: which files does the Go build **not** need?
5. **Build** it (from the project root folder):
   ```bash
   docker build -t order-service ./order-service
   ```
6. **Run it alone** (no database):
   ```bash
   docker run --rm order-service
   ```
   ✅ Expected: it stops with the error `DATABASE_URL is required`.
   This is **good**! It shows the binary starts and reads the config. (The database comes in DS-5.)
7. **Compare sizes:**
   ```bash
   docker pull golang:1.27
   docker images
   ```
   Write down the size of `order-service` and of `golang`.
8. **See the cache work:** change only a comment in `main.go`, then build again.
   Look at the output: which steps say `CACHED`? Is `go mod download` cached? Why?
9. Push your work or paste it to Claude → Claude reviews it.

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

### Words you need

| Word               | Meaning                                                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| **Docker Compose** | A tool to start **many containers together** from one YAML file.                                                          |
| **Service**        | One container in the compose file (here: `postgres` and `order-service`).                                                 |
| **Network**        | Compose puts all services in one private network. They find each other **by service name**.                               |
| **Named volume**   | A storage place managed by Docker. It lives **outside** the container, so data is not lost when the container is deleted. |
| **Bind mount**     | A file or folder from **your computer**, shown inside the container (we use it for `init.sql`).                           |
| **Healthcheck**    | A command Docker runs again and again to ask "are you ready?". The answer is `healthy` or `unhealthy`.                    |
| **`depends_on`**   | "Start this service after that one." With `condition: service_healthy` → "after that one is **ready**".                   |

```
your computer
  │  localhost:8080
  ▼
┌──────────── compose network ────────────┐
│  order-service  ──postgres:5432──►  postgres  │
│                                     │    │
└─────────────────────────────────────┼────┘
                                      ▼
                              named volume (data)
```

### Steps to do

1. Create `docker-compose.yml` in the project root folder.
2. Add the **`postgres` service**:
   1. `image` — the official Postgres image.
   2. `environment` — user, password, database name. Choose simple values and **write them down**. You need them again in step 4.
   3. `volumes` — two lines: the named volume for data, and your `init.sql` file into `/docker-entrypoint-initdb.d/`.
   4. `healthcheck` — use `pg_isready`.
3. Add the **`volumes:`** section at the bottom (the top level of the file), and name your volume.
4. Add the **`order-service` service**:
   1. `build` — the folder with the Dockerfile.
   2. `ports` — make `8080` reachable from your computer.
   3. `environment` — `DATABASE_URL`. Use the same user / password / DB name as in step 2. The host is the **service name**.
   4. `depends_on` — wait for Postgres to be **healthy**.
5. **Start everything:**
   ```bash
   docker compose up --build -d
   docker compose ps
   ```
   ✅ Expected: `postgres` is `healthy`, and `order-service` is `running`.
6. **Read the logs:**
   ```bash
   docker compose logs order-service
   ```
   ✅ Expected: `connected to database` and `listening`.
7. **Test the API:**
   ```bash
   curl -i http://localhost:8080/health
   curl -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":"u1","amount":1999}'
   curl -i http://localhost:8080/orders/<id-from-above>
   ```
8. **Test bad input:** send `amount: -5`, and a wrong id like `/orders/abc`. Which status codes do you get?
9. Stop everything: `docker compose down` (the data stays in the volume).
10. Push your work or paste it to Claude → Claude reviews it.

**If something goes wrong:**

- `password authentication failed` → user/password in `DATABASE_URL` and in `postgres` are different.
- `relation "orders" does not exist` → `init.sql` did not run. Maybe the volume already had data from an earlier try.
  Run `docker compose down -v` (⚠️ this deletes the data) and start again.
- `connection refused` → wrong host or port in `DATABASE_URL`.

**Acceptance criteria**

- [ ] A `postgres` service: official image (for example `postgres:17`), with user, password and database name set by env vars.
- [ ] Your `init.sql` (DS-3) is mounted, so Postgres runs it on the first start.
- [ ] Postgres data is in a **named volume**, so it is not lost after `docker compose down`.
- [ ] Postgres has a **healthcheck**.
- [ ] An `order-service` service: builds from `./order-service`, publishes port `8080`, and gets `DATABASE_URL` as an env var.
- [ ] `order-service` waits until Postgres is **healthy** (ready to work), not only **started**.
- [ ] `curl http://localhost:8080/health` returns `200`.
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

Test commands:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":"u1","amount":1999}'
curl -i http://localhost:8080/orders/<id-from-above>
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

<details>
<summary>Hint 1</summary>

- Measure the time of a request: `curl -w "%{time_total}\n" ...`
- `pgxpool` opens connections **when they are needed**. What does this mean when Postgres comes back?
- At start, `main.go` checks the DB and stops if it fails. Compare this with "the DB stops while the service is running".
</details>

---

## DS-7 — Install Ubuntu in WSL

**Type:** Task · **Who:** 👤 Me · **Status:** ✅ Done

**Description**
WSL2 works, but `wsl -l -v` shows only `docker-desktop`. This is Docker's own small Linux. It is not for daily work.
Install Ubuntu, so you can follow the Terminal rule. This is the last ticket that starts in PowerShell.

**Acceptance criteria**

- [x] `wsl -l -v` shows `Ubuntu` with version `2`, and a `*` next to it (it is the default).
- [x] With Docker Desktop open, `docker version` in Ubuntu shows both **Client** and **Server**.
- [x] In Ubuntu, I can go to the project folder and `git status` works.
- [x] `go version`, `curl --version` and `jq --version` work in Ubuntu.
- [x] I can answer: What is `/mnt/c`? Where does the `docker` command send its work?

<details>
<summary>Hint 1</summary>

- Install from PowerShell with `wsl --install`. Search how to choose a distribution with `-d`.
- Connect Docker: Docker Desktop → Settings → Resources → WSL integration → turn on Ubuntu.
- `curl`, `jq` and `git` come from `apt`.
- Get Go from go.dev/dl. The Ubuntu `apt` version can be too old for our `go.mod`.
- Git says `dubious ownership`? Read the message. It gives you the command to run.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```bash
# in PowerShell
wsl --install -d ____

# in Ubuntu
sudo apt update
sudo apt install -y ____ ____ ____

# Go: download the Linux file from go.dev/dl first
sudo tar -C /usr/local -xzf ____.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

</details>

---

## DS-8 — Linux: files and search

**Type:** Task · **Who:** 👤 Me · **Status:** ✅ Done · **Needs:** DS-7

**Description**
Learn to move around and find things with only the keyboard. You use these commands every day on a server.

**Acceptance criteria**

- [x] In `~/practice`, I create folders, then create, read, copy, move and delete a file. I use `pwd`, `ls -la`, `cd`, `mkdir -p`, `cat`, `less`, `cp`, `mv` and `rm`.
- [ ] I find every place that uses `DATABASE_URL`, with file name and line number.
- [ ] I find all `.go` files in the project.
- [x] I can answer: What do the columns in `ls -la` mean? What does `~` mean? Can I get a file back after `rm`?

<details>
<summary>Hint 1</summary>

- `man grep` shows the help for a command. Press `q` to leave. Press `q` to leave `less` too.
- `-r` means "look inside folders". `-n` means "show line numbers".
- Too many results from `.git`? Search for the `--exclude-dir` option.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```bash
grep -rn ____ . --exclude-dir=.git
find . -name "____"
```

</details>

---

## DS-9 — Experiment: container vs volume

**Type:** Spike · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-5

**Description**
Learn where your data really lives. Write what you see in `notes/step1.md`: 2–3 lines with what you did, what you saw, and why.

**Acceptance criteria**

- [ ] I create an order and save its id.
- [ ] I run `docker compose down`, then `up`. I read the order. I write down the result.
- [ ] I run `docker compose down -v`, then `up`. I read the order. I write down the result.
- [ ] I can answer: What did `-v` delete? Did `init.sql` run again? Why?

<details>
<summary>Hint 1</summary>

- `docker volume ls` shows the volumes. Run it after each `down`.
- Read DS-3 again: when does Postgres run `init.sql`?
</details>

---

## DS-10 — Linux: env, processes and ports

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-7

**Description**
Run the service directly in Ubuntu, with no container. Give it its config with environment variables, then find its process and its port.

**Steps to do**

1. Start a temporary Postgres. This one publishes a port, so Ubuntu can reach it:
   ```bash
   docker run -d --name pg-test -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17
   ```
   Why `55432`? Windows on this computer has its own Postgres on ports 5432 and 5433.
   If you use those ports, you can connect to the wrong database and get `password authentication failed`.
2. In `order-service/`, run `go run .` with no variables. Read the error.
3. Run it again with `DATABASE_URL`, and with `PORT=8081`, so it does not use the same port as Compose.
4. In a second terminal, look for the process and the port.
5. Clean up: stop the service with Ctrl+C, then run `docker rm -f pg-test`.

**Acceptance criteria**

- [ ] I can show a variable with `env` and `echo $HOME`, and I can set one with `export`.
- [ ] `curl http://localhost:8081/health` returns `200`.
- [ ] I found the process with `ps aux`, and watched it in `top`.
- [ ] `ss -tlnp` shows which program listens on port 8081.
- [ ] I can answer: What is the difference between `VAR=x command` and `export VAR=x`? What is the program name, and why is it not `go`?

<details>
<summary>Hint 1</summary>

- Connection URL: `postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable`.
- The default user and database in the Postgres image are both `postgres`.
- From Ubuntu, the host is `localhost`, and the port is the left number in `-p 55432:5432`.
- You do not see the program name in `ss`? Try `sudo ss -tlnp`.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```bash
PORT=____ DATABASE_URL="postgres://postgres:____@localhost:____/postgres?sslmode=disable" go run .

ps aux | grep ____
sudo ss -tlnp | grep ____
```

</details>

---

## DS-11 — Linux: permissions and pipes

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-5

**Description**
Learn to read file permissions and to connect small commands together.

⚠️ Do the permissions practice in your Linux home folder (`~`), not in `/mnt/c`.
Windows files do not normally keep Linux permissions, so `chmod` may seem to do nothing there.

**Acceptance criteria**

- [ ] I can read `rwxr-xr-x`: who can read, write and run the file?
- [ ] I create `~/practice/hello.sh`. `./hello.sh` fails with `Permission denied`. After `chmod +x`, it works.
- [ ] I stop Postgres and make a few requests, so the logs have errors. Then I run `docker compose logs order-service | grep -i error | wc -l`.
- [ ] I can explain each of the 3 parts of this pipe, and what `|` does.
- [ ] I can answer: Why must a container not run as `root`?

<details>
<summary>Hint 1</summary>

- Read the 9 letters as 3 groups: owner, group, everyone else.
- `r` = read, `w` = write, `x` = run.
- A pipe `|` sends the output of the left command to the input of the right command.
- `man wc`: what does `-l` count?
</details>

---

## DS-12 — Linux: look inside a container

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-5

**Description**
See with your own eyes that a container is only a Linux process with its own view of the system.

**Acceptance criteria**

- [ ] I open a shell in the Postgres container with `docker exec -it <postgres-container> bash`.
- [ ] I run `ps aux` and `cat /etc/os-release` inside it.
- [ ] I write down: how many processes are there? Which program is PID 1? Which Linux is it? Is it the same as Ubuntu?
- [ ] I compare the count with `ps aux | wc -l` in Ubuntu.
- [ ] I can answer: Why so few processes? What happens to the container when PID 1 stops?

<details>
<summary>Hint 1</summary>

- `docker compose ps` shows the container name.
- Search for "Linux PID namespace".
</details>

---

## DS-13 — Bash script: `smoke-test.sh`

**Type:** Task · **Who:** 👤 Me · **Status:** 🔲 To Do · **Needs:** DS-5, DS-11

**Description**
A **smoke test** is a short test that answers one question: "Does the main path work?"
Write it once. You use it again in CI, on the AWS server, and after every deploy.

**File:** `scripts/smoke-test.sh`

**Acceptance criteria**

- [ ] The script calls `/health` and expects `200`.
- [ ] It creates an order and expects `201`.
- [ ] It reads the order by its id, expects `200`, and checks that the id is the same.
- [ ] It takes the address as its first argument, and uses `http://localhost:8080` if there is none.
- [ ] It **exits with code 1** if something fails, and with code 0 if all is good. It prints which check failed.
- [ ] With Postgres stopped, `scripts/smoke-test.sh; echo $?` prints `1`. With the service stopped, it also prints `1`.
- [ ] Git saves the file as executable: `git ls-files --stage scripts/smoke-test.sh` starts with `100755`.
- [ ] I can answer: What is an exit code? How does CI use it? What does `set -euo pipefail` do?

<details>
<summary>Hint 1</summary>

- Start the file with `#!/usr/bin/env bash`.
- Get only the status code: search for the `curl` options `-o /dev/null` and `-w`.
- `jq -r .id` takes the `id` out of the JSON.
- `${1:-something}` means "the first argument, or `something` if there is none".
- This repo is on the Windows disk, and Git ignores `chmod` here. Use `git add --chmod=+x`.
- Error `bad interpreter` or `^M`? The file has Windows line endings. In VS Code, click `CRLF` at the bottom and choose `LF`.
</details>

<details>
<summary>Hint 2</summary>

Fill in the blanks:

```bash
#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-____}"

fail() {
  echo "FAIL: $1" >&2
  exit ____
}

# 1. health
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/____" || true)
[ "$code" = "____" ] || fail "health returned $code"

# 2. create an order: save the body to a file, and keep the status code
code=$(curl -s -o /tmp/order.json -w "%{http_code}" -X POST "$BASE_URL/orders" \
  -H "____" -d '____' || true)
[ "$code" = "____" ] || fail "create returned $code"
id=$(jq -r ____ /tmp/order.json)

# 3. read it back: write this part yourself

echo "OK"
```

</details>

---

## Definition of Done (Step 1)

- [ ] DS-1 … DS-13 are done. Claude reviewed DS-3, DS-4, DS-5 and DS-13 from what I pushed or pasted.
- [ ] `notes/step1.md` has my experiment results.
- [ ] I can explain without looking: request → handler → DB → response; what a connection pool is and why we need it;
      why multi-stage builds make images smaller; how containers find each other by name in Compose;
      a container is just a Linux process; where container data lives (volumes); what an exit code is and why scripts need it.
- [ ] Committed and pushed to git.
