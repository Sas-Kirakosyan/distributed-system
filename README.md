# Learning Plan: Orders & Payments — A Small Distributed System

**Goal:** Build a small order/payment system in thin layers. Each step adds one new concept, and I don't move on until I can explain every part of it out loud.

**Pace:** ~1 hour per day. Time estimates are rough; understanding beats speed.

**Rule for every step:** It's done when I can answer these without looking at the code:

1. What does each part do, and how does a request flow through the system?
2. Why is it built this way? What's the alternative?
3. What breaks first, and what happens when it does?

---

## Working agreement (who does what)

| Part | Who | How |
|---|---|---|
| **Go application code** | Claude | Writes it and explains it. I read and understand every line. |
| **Dockerfiles, Compose, SQL scripts, env config** | **Me** | Claude gives me the task + hints, never the full answer. |
| **Kafka / Prometheus / Grafana / k6 / CI / Kubernetes config** | **Me** | Same: task + hints. |
| **Running experiments** | **Me** | I break things and write down what I saw. |
| **Testing & verification** | Claude | Checks my work via the terminal and tells me what's wrong — and why. |

Tasks marked 👤 are mine. Tasks marked 🤖 are Claude's.

---

## Step 1 — One service, one database (~1.5–2 weeks)

**New concept:** A backend service, persistence, and containers.

- [x] Set up WSL2 + Docker Desktop.
- [x] Order Service in Go (Gin): `POST /orders` and `GET /orders/:id`.
- [ ] PostgreSQL with an `orders` table (`id`, `user_id`, `amount`, `status`, `created_at`), created by a plain SQL init script.
- [x] Connect with `pgxpool`; read config (DB URL, port) from environment variables.
- [x] Basic input validation and proper HTTP status codes (400, 404, 500).
- [x] A `/health` endpoint.
- [ ] Multi-stage Dockerfile.
- [ ] `docker-compose.yml` with Order Service + Postgres.

**Experiments:**

- [ ] Stop Postgres while the service runs. What does the API return? What do the logs say?
- [ ] Restart Postgres. Does the service recover on its own?

📋 **Tickets with acceptance criteria and hints: [tasks/STEP-1.md](tasks/STEP-1.md)**

**Done when I can explain:** request → handler → DB → response; what a connection pool is and why it exists; why multi-stage builds make images smaller; how containers find each other by name in Compose.

> **Interview talking point:** "I containerized a Go service with Postgres, configured it via environment variables, and tested what happens when the database goes down."

---

## Step 2 — Two services talking over HTTP (~1.5–2 weeks)

**New concept:** Service-to-service communication and partial failure — the core distributed systems problem.

- [ ] Payment Service in Go: `POST /payments` that marks a payment as succeeded or failed (mock logic, e.g. fail if `amount > 1000`).
- [ ] Payment Service has its own Postgres database (or own schema) with a `payments` table. Each service owns its data.
- [ ] Order Service calls Payment Service synchronously after saving the order, then updates the order status (`PAID` / `PAYMENT_FAILED`).
- [ ] Add a timeout to the HTTP client.
- [ ] Add both services to Compose.

**Experiments:**

- [ ] Stop Payment Service. What happens to new orders?
- [ ] Add an artificial 5-second delay in Payment Service. What happens to Order Service's response time?
- [ ] Add a simple retry. Can a retry cause a double payment? *(Answer: yes — remember this, it comes back in Step 3.)*

**Done when I can explain:** why each service has its own database; what happens when a downstream service is slow or down; why timeouts are mandatory; why retries are dangerous without idempotency; what "tight coupling" means here.

> **Interview talking point:** "I started with synchronous HTTP between services, then deliberately broke the payment service to see how failure spreads — that's what motivated moving to async messaging."

---

## Step 3 — Replace HTTP with Kafka (~2–3 weeks)

**New concept:** Asynchronous, event-driven communication.

- [ ] Add Kafka to Compose (single broker, KRaft mode — no Zookeeper).
- [ ] Order Service saves the order as `PENDING`, publishes an `OrderCreated` event, and returns `202 Accepted` immediately.
- [ ] Payment Service consumes `OrderCreated` (`segmentio/kafka-go`), processes payment, and publishes `PaymentCompleted` or `PaymentFailed`.
- [ ] Order Service consumes payment events and updates order status.
- [ ] Use `order_id` as the message key.

**Experiments:**

- [ ] Stop Payment Service, create 10 orders, start it again. What happens? (Compare with Step 2.)
- [ ] Make Payment Service crash after processing but before committing the offset. Does the same order get paid twice?
- [ ] Make Payment Service idempotent: store processed `order_id`s and skip duplicates.
- [ ] Kill Kafka after saving an order but before publishing. The order exists but no event was sent — this is the **dual-write problem**.
- [ ] *Optional (only if time allows):* Fix the dual-write problem with the **Transactional Outbox** pattern.

**Done when I can explain:** sync vs async trade-offs; topics, partitions, keys, consumer groups and offsets in simple terms; at-least-once delivery and why duplicates happen; idempotency; the dual-write problem and how the outbox solves it; why the API returns 202 instead of 200.

> **Interview talking point:** "Moving to Kafka made the system resilient to payment outages, but introduced duplicate messages, so I made the consumer idempotent. I also found the dual-write problem between Postgres and Kafka and know the outbox pattern fixes it."

> ⭐ **Steps 1–3 are the core.** If I stop here, I already have a strong, fully explainable project.

---

## Step 4 — Observe and break it (~1.5–2 weeks)

**New concept:** Measuring a system under load.

- [ ] Add Prometheus metrics to both services (`/metrics`): request count, error count, latency histogram.
- [ ] Add Prometheus + Grafana to Compose; build one simple dashboard.
- [ ] Write a k6 script for `POST /orders`: ramp up, hold, ramp down.
- [ ] Run it, find the first bottleneck, change one thing, run again, compare.

**Done when I can explain:** p50 vs p95 vs p99 latency and why averages lie; what the bottleneck was and how I proved it; consumer lag and what it tells me.

> **Interview talking point:** "I load-tested with k6, saw p99 latency spike at X requests/sec, traced it to Y using Grafana, fixed it, and measured the improvement."

---

## Step 5 — CI/CD (~1 week)

**New concept:** Automating build and release.

- [ ] GitHub Actions: on push to `main` → `go test`, lint, build.
- [ ] Build Docker images and push to GitHub Container Registry (GHCR).

**Done when I can explain:** what each pipeline step does; how secrets are handled; why tests run before the image is built.

---

## Step 6 — Kubernetes (later, ~2–3 weeks)

**New concept:** Orchestration.

- [ ] Minikube; Deployments + Services for both Go services.
- [ ] ConfigMaps and Secrets for configuration.
- [ ] Readiness and liveness probes (using `/health`).
- [ ] Scale Payment Service to 3 replicas and watch Kafka consumer groups split the partitions.

**Done when I can explain:** Pod vs Deployment vs Service; why probes matter for zero-downtime deploys; how scaling interacts with Kafka partitions.

---

## Deliberately left out (for now)

- **Redis** — adds a cache-vs-source-of-truth problem before the basics are solid.
- **AWS / cloud** — local first.
- **Helm, Terraform, service mesh** — only after Step 6, if ever.
