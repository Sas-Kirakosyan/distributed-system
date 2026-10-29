# How the Order Service talks to Postgres

## Short answers

| Question | Answer |
|---|---|
| Do we use an ORM? | **No.** We write SQL by hand. |
| What package do we use? | **pgx** (`github.com/jackc/pgx/v5`). It is a Postgres **driver**. |
| One connection or many? | Many. We use a **connection pool** (`pgxpool`). |
| Do we need to log in (authenticate)? | **Yes.** Postgres always checks a username and password. |
| Is the connection encrypted? | **Not on your laptop** (`sslmode=disable`). In the cloud it must be. |

---

## 0. Concept first, tool second (most important part)

**Tools change. Concepts stay.**
You may forget the name `pgxpool` in one year. That is OK.
But you must remember **why** a pool exists. The concept is the root. The tool is only one way to do it.

When you learn a new tool, always ask:

1. **What problem does it solve?**
2. **What is the general idea (the concept) behind it?**
3. **How would it look in another language or tool?**

If you can answer these, you can learn any new tool fast.

| Tool we use (can forget) | Concept behind it (must remember) | Same concept in other places |
|---|---|---|
| `pgx` | **Database driver** — speaks the DB network protocol | JDBC (Java), `psycopg` (Python), `pg` (Node.js) |
| `pgxpool` | **Connection pooling** — reuse open connections, because opening one is slow | HikariCP (Java), PgBouncer, HTTP keep-alive |
| `$1, $2` | **Parameterized queries** — send SQL and values separately → no SQL injection | `?` in MySQL/JDBC, `%s` in psycopg |
| `DATABASE_URL` | **Config from the environment** — no settings or secrets in code | 12-Factor App, Kubernetes ConfigMap/Secret |
| `sslmode` | **Encryption in transit (TLS)** — nobody on the network can read the data | HTTPS for websites |
| `context` with timeout | **Timeouts** — never wait forever for another system | timeouts in every HTTP client and DB client |
| `repository.go` | **Separation of layers** — only one place knows about SQL | Repository pattern in any language |
| `POSTGRES_USER` superuser | **Least privilege** — give only the rights that are needed | IAM roles in the cloud, Linux users |

**In interviews**, people ask about concepts much more than about tools.
"How does a connection pool work?" is a common question. "What is the API of pgxpool?" is not.

---

## 1. The big picture

```
HTTP request
    │
    ▼
handler.go        checks the input, sets a 3-second timeout
    │
    ▼
repository.go     has the SQL text, calls the pool
    │
    ▼
pgxpool           gives us a free connection (or opens a new one)
    │
    ▼  TCP, port 5432, with username + password
    ▼
Postgres          runs the SQL, sends rows back
```

Only `repository.go` knows about SQL. The handler does not know about SQL.
This makes the code easy to read and easy to change later.

---

## 2. Driver vs ORM

**Driver** = a library that knows how to "speak" the Postgres network protocol.
You give it SQL text, it sends it to Postgres and gives you the result.

**ORM** (for example GORM in Go) = a bigger library that sits on top of a driver.
It writes the SQL **for you** from your Go structs.

```go
// ORM style (GORM) — you don't see the SQL
db.Create(&order)

// Our style (pgx) — you see the exact SQL
pool.QueryRow(ctx, "INSERT INTO orders (user_id, amount, status) VALUES ($1, $2, $3) RETURNING ...", ...)
```

**Why we don't use an ORM here:**

- **Learning.** You see every SQL query. You learn SQL, not a library.
- **Small app.** We have only 2 queries. An ORM is too much for that.
- **Control.** You know exactly what runs on the database. ORMs sometimes make slow queries you don't see.
- **Common in Go.** Many Go teams use plain SQL or `sqlc` instead of an ORM.

**The bad side of no ORM:**

- More code to write by hand (the `Scan(&o.ID, &o.UserID, ...)` lines).
- A typo in SQL is found only when the code runs, not when it compiles.

**Middle option for later:** `sqlc`. You write SQL, and it **generates** the Go code for you.
Good to know for interviews.

**Why pgx and not `database/sql`?**
`database/sql` is Go's standard, generic database interface. It works with any database.
`pgx` is made only for Postgres. It is faster, and it supports Postgres types better.
For a Postgres-only project, pgx is the usual choice today.

---

## 3. The connection URL

The service gets **one** setting for the database: the `DATABASE_URL` environment variable.

```
postgres://orders_user:secret@postgres:5432/orders?sslmode=disable
```

| Part | Meaning |
|---|---|
| `postgres://` | the protocol |
| `orders_user` | **username** |
| `secret` | **password** |
| `postgres` | **host** — the Compose service name (not `localhost`!) |
| `5432` | the Postgres port |
| `orders` | the **database name** |
| `sslmode=disable` | no encryption (OK only on your laptop) |

The code reads it in [config.go](../order-service/internal/config/config.go).
If `DATABASE_URL` is missing, the service stops at start with an error. This is on purpose:
it is better to fail early and loudly than to start a broken service.

**Why the host is not `localhost`:** inside a container, `localhost` means "this container".
Postgres runs in a *different* container. Compose gives every service a name you can use
like a hostname. So the host is `postgres` (the service name).

---

## 4. Authentication (logging in)

**Yes, we authenticate.** Postgres always asks "who are you?" and checks the password.

**Where does the user come from?**
The official Postgres Docker image creates a user on its **first** start, from these env vars:

```yaml
POSTGRES_USER: orders_user
POSTGRES_PASSWORD: secret
POSTGRES_DB: orders
```

The same username and password must be in `DATABASE_URL`. If they don't match, the service
can't connect, and you see an error like `password authentication failed for user ...`.

**Rules for passwords (remember these):**

1. **Never put a password in Go code.** Always read it from an environment variable.
2. **Never commit a real password to git.** For local learning, a simple password in
   `docker-compose.yml` is OK. Better: put it in a `.env` file. Compose reads `.env`
   automatically, and our `.gitignore` already ignores `.env`.
3. **Never use `POSTGRES_HOST_AUTH_METHOD=trust`.** It turns off the password check completely.
4. In AWS (Step 5), passwords go into **SSM Parameter Store**. In Kubernetes (Step 8), they go into a **Secret**.

**Two more safety points (good for interviews):**

- **Least privilege.** `POSTGRES_USER` is a *superuser* (it can do anything). In a real
  company, the app uses a separate user that can only read and write its own tables.
  We skip this for now, but know that it exists.
- **Don't open the port if you don't need it.** If Compose publishes `5432:5432`, then other
  programs on your computer (or network) can try to connect. The Order Service does **not**
  need this port to be published; it talks to Postgres inside the Compose network.
  Publish it only if you want to connect with a tool like `psql` or DBeaver, and then use
  `127.0.0.1:5432:5432` so only your own computer can reach it.

**Encryption (`sslmode`):**
`sslmode=disable` means the data (and the password!) travels without encryption.
This is OK on your laptop, because the traffic never leaves your computer.
In the cloud (RDS in Step 5), use `sslmode=require` or, better, `sslmode=verify-full`.

---

## 5. The connection pool

**Problem:** opening a new connection to Postgres is **slow**. It needs a TCP connection,
the login check, and Postgres starts a new process for each connection. If we open a new
connection for every HTTP request, the service becomes slow.

**Solution:** a **pool** = a group of connections that stay open and are **reused**.

```
Request 1 ──► borrow conn A ──► run SQL ──► give conn A back
Request 2 ──► borrow conn B ──► run SQL ──► give conn B back
Request 3 ──► borrow conn A (free again) ──► ...
```

**How our code uses it** ([main.go](../order-service/main.go)):

1. `pgxpool.New(...)` — creates the pool. **It does not connect yet.**
2. `pool.Ping(...)` — opens a real connection to check the DB works. If not, the service stops.
3. `pool.QueryRow(...)` — borrows a connection, runs the SQL, gives the connection back. All automatic.
4. `pool.Close()` — closes all connections when the service stops.

**Pool size:**
By default, pgx allows `max(4, number of CPUs)` connections. You can change it in the URL,
for example `...?sslmode=disable&pool_max_conns=10`.

Postgres also has a limit: `max_connections` = 100 by default.
So if you run 5 copies of the service, each with 30 connections, that is 150 → too many!
Remember this for Step 8 (Kubernetes), when we start more copies.

**What if Postgres goes down and comes back?** Try it in DS-6 before you read about it. 😉

---

## 6. One request, step by step

`POST /orders` with body `{"user_id": "u1", "amount": 1999}`:

1. Gin calls `create()` in [handler.go](../order-service/internal/order/handler.go).
2. `ShouldBindJSON` reads the JSON and checks the rules (`user_id` is required, `amount` > 0).
   If something is wrong → **400**.
3. We make a `context` with a **3-second timeout**. If the DB is slow, we stop waiting after 3 s.
4. The handler calls `repo.Create(...)` in [repository.go](../order-service/internal/order/repository.go).
5. `pool.QueryRow` borrows a connection and sends the `INSERT` with the values `$1, $2, $3`.
6. Postgres saves the row, makes the `id` and `created_at`, and sends the row back (`RETURNING ...`).
7. `Scan(...)` copies the row into the Go `Order` struct. The connection goes back to the pool.
8. The handler sends **201 Created** with the order as JSON.
   If any DB error happened → the error is written to the log, and the client gets a simple **500**.

---

## 7. Why `$1, $2` and not string building? (SQL injection)

**Wrong — never do this:**

```go
q := fmt.Sprintf("SELECT * FROM orders WHERE id = '%s'", id)
```

If someone sends `id = x' OR '1'='1`, the query changes and returns **all** orders.
This attack is called **SQL injection**.

**Right — what we do:**

```go
pool.QueryRow(ctx, "SELECT ... FROM orders WHERE id = $1", id)
```

With `$1`, the SQL text and the values are sent to Postgres **separately**.
Postgres never treats the value as SQL code. So SQL injection is not possible.

---

## Check yourself (answer without looking)

- [ ] What is the difference between a driver and an ORM? Why no ORM here?
- [ ] What are the parts of `DATABASE_URL`? Why is the host not `localhost`?
- [ ] Where does the Postgres user come from? What happens if the password is wrong?
- [ ] Why should the password never be in Go code or in git?
- [ ] What does `sslmode=disable` mean? When is it OK, and when is it not OK?
- [ ] What is a connection pool, and why do we need one?
- [ ] Does `pgxpool.New` connect to the database? What does?
- [ ] Why `$1` instead of `fmt.Sprintf`?
