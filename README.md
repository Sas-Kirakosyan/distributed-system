# Learning Plan: DevOps with a Small Orders & Payments System

**Main goal:** learn **DevOps skills** by doing them by hand:
Linux, Docker, CI/CD, AWS, Terraform, monitoring and Kubernetes.

The Go app is **small on purpose**. It is only "the thing we deploy".
The real work is everything around it: **build it, ship it, run it, watch it, fix it.**

**Target:** a project I can show and explain in a **DevOps / Cloud / Platform engineer** interview.

**Pace:** ~1 hour per day. Time estimates are rough; understanding beats speed.

**Rule for every step:** It's done when I can answer these without looking:

1. What does each part do, and how does a request flow through the system?
2. Why is it built this way? What's the alternative?
3. What breaks first, and what happens when it does?
4. Can I delete everything and build it again from zero? How long does it take?

---

## Skills map

**Concept first, tool second.** Tools change. Concepts stay.

| Area | Concept (must remember) | Tools we use (can change) | Steps |
|---|---|---|---|
| **Linux** | files, permissions, users, processes, ports, services, logs | bash, ssh, systemd, journalctl | all steps (most in 1 and 3) |
| **Containers** | pack an app with everything it needs; a container is a Linux process | Docker, Compose | 1 |
| **CI** | every push is tested and built by a machine, not by hand | GitHub Actions | 2 |
| **Cloud** | rent servers, networks and managed services; pay per hour | AWS (IAM, EC2, VPC, ECR, RDS, CloudWatch) | 3, 5 |
| **Infrastructure as Code (IaC)** | infra is written in files, reviewed in git, and rebuilt with one command | Terraform | 4 and after |
| **CD** | deploys are automatic, repeatable, and easy to roll back | GitHub Actions + AWS | 5 |
| **Security** | least privilege, no secrets in code, short-lived credentials | IAM roles, OIDC, SSM Parameter Store | 3–5 |
| **Observability** | metrics, logs, alerts: see the problem before users tell you | Prometheus, Grafana, CloudWatch, k6 | 6 |
| **Distributed systems** | services fail separately; async messages; idempotency | SQS (Kafka is the same idea) | 7 |
| **Orchestration** | you describe the wanted state; the system keeps it by itself | Kubernetes, Helm | 8 |

---

## Working agreement (who does what)

| Part | Who | How |
|---|---|---|
| **Go application code + Go tests** | Claude | Writes it and explains it. I read and understand every line. |
| **Linux work, bash scripts** | **Me** | Claude gives me the task + hints, never the full answer. |
| **Dockerfiles, Compose, SQL scripts, env config** | **Me** | Same: task + hints. |
| **GitHub Actions, AWS, Terraform** | **Me** | Same: task + hints. |
| **Prometheus / Grafana / k6 / Kubernetes / Helm config** | **Me** | Same: task + hints. |
| **Running experiments** | **Me** | I break things and write down what I saw. |
| **Testing & verification** | Claude | Checks my work via the terminal and tells me what's wrong — and why. |

Tasks marked 👤 are mine. Tasks marked 🤖 are Claude's. 🐧 = Linux practice.

**Writing rule:** All docs, tasks and explanations use **simple B1-level English**:
short sentences, common words, one idea per sentence. Explain technical words the first time they appear.

**Terminal rule:** From now on, I work in the **Ubuntu (WSL) terminal**, not in PowerShell.
Real servers run Linux. Every command I type there is practice.

📚 **Docs:**
- [docs/database.md](docs/database.md) — how the service connects to Postgres (no ORM, connection pool, login, security)

---

## 💰 AWS cost safety (read before Step 3)

The cloud costs **real money**. A forgotten server costs money **every hour**, even when nobody uses it.

**Before I create anything in AWS:**
- [ ] Turn on **MFA** for the root user. Never use the root user for daily work.
- [ ] Create an admin user for myself (IAM Identity Center or an IAM user) and use only that.
- [ ] Create a **budget alert** in AWS Budgets (for example $10/month) that sends me an email.
- [ ] Choose **one region** and always use it. (Things in other regions are easy to forget.)
- [ ] Check what my account gets for free. AWS changed its free tier in 2025, so read the current rules.

**Rules for every session:**
- End every session with **nothing running**. After Step 4 this is one command: `terraform destroy`.
- Use the smallest sizes (for example `t3.micro` or `t4g.micro`).
- Watch out for things that cost money even when idle:
  **NAT Gateway** (~$30+/month), **EKS** control plane (~$73/month), **MSK** (managed Kafka),
  load balancers, RDS databases, and **public IPv4 addresses** (AWS charges for each one).
- Check **Billing → Bills** the day after each session.

---

## Step 1 — Linux + Docker: one service, one database (~2 weeks) 🔨 in progress

**New concept:** containers, and the Linux command line.

- [x] Set up WSL2 + Docker Desktop.
- [x] 🤖 Order Service in Go (Gin): `POST /orders` and `GET /orders/:id`.
- [ ] PostgreSQL with an `orders` table (`id`, `user_id`, `amount`, `status`, `created_at`), created by a plain SQL init script.
- [x] 🤖 Connect with `pgxpool`; read config (DB URL, port) from environment variables.
- [x] 🤖 Basic input validation and proper HTTP status codes (400, 404, 500).
- [x] 🤖 A `/health` endpoint.
- [ ] Multi-stage Dockerfile.
- [ ] `docker-compose.yml` with Order Service + Postgres.

**🐧 Linux practice** (in the Ubuntu WSL terminal):

- [ ] Move around and work with files: `pwd`, `ls -la`, `cd`, `mkdir -p`, `cat`, `less`, `cp`, `mv`, `rm`.
- [ ] Search: find every place that uses `DATABASE_URL` with `grep -rn`. Find all `.go` files with `find`.
- [ ] Environment variables: `env`, `export`, `echo $HOME`. Run the service with `DATABASE_URL=... go run .`
- [ ] Processes and ports: `ps aux`, `top`, `ss -tlnp`. Which program listens on port 8080?
- [ ] Permissions: what does `rwxr-xr-x` mean? Use `chmod +x`. Why must a container not run as `root`?
- [ ] Pipes: `docker compose logs order-service | grep -i error | wc -l`. What does each part do?
- [ ] Look inside a container: `docker exec -it <postgres-container> bash`, then `ps aux` and `cat /etc/os-release`.
  How many processes do you see? Why so few?
- [ ] Write my first bash script `scripts/smoke-test.sh`: it calls `/health`, creates an order, reads it back,
  and **exits with code 1** if something fails. (I use this script again in CI and after every deploy.)

**Experiments:**

- [ ] Stop Postgres while the service runs. What does the API return? What do the logs say?
- [ ] Restart Postgres. Does the service recover on its own?

📋 **Tickets with acceptance criteria and hints: [tasks/STEP-1.md](tasks/STEP-1.md)**

**Done when I can explain:** request → handler → DB → response; what a connection pool is and why it exists;
why multi-stage builds make images smaller; how containers find each other by name in Compose;
**a container is just a Linux process** with its own view of files, network and other processes;
what an exit code is and why scripts need it.

> **Interview talking point:** "I containerized a Go service with Postgres, configured it via environment variables, wrote a smoke-test script, and tested what happens when the database goes down."

---

## Step 2 — CI: test and build on every push (~1 week)

**New concept:** Continuous Integration — a machine checks every change, not me.

- [ ] Push the repo to GitHub.
- [ ] 🤖 A few Go tests (`go test ./...`).
- [ ] GitHub Actions workflow on every push and pull request: `go vet`, `go test`, lint, `docker build`.
- [ ] Cache Go modules and Docker layers, so the pipeline is fast.
- [ ] On push to `main`: push the image to GitHub Container Registry (GHCR).
  Tag it with the **git commit SHA**, not only `latest`.
- [ ] Scan the image for known security problems (for example with Trivy).
- [ ] Start Compose inside CI and run `scripts/smoke-test.sh` (an integration test).
- [ ] Branch protection: nobody can merge to `main` if CI is red.

**Experiments:**

- [ ] Break a test on purpose and open a pull request. Is the merge blocked?
- [ ] Change only a comment. How much faster is the pipeline with the cache?

**Done when I can explain:** CI vs CD; why we tag images with the commit SHA and not `latest`;
how secrets work in GitHub Actions; why tests run before the image is built; why a pipeline should fail fast.

> **Interview talking point:** "Every push runs tests, a security scan and an integration test with Docker Compose. Images are tagged with the commit SHA, so I always know which code runs where."

---

## Step 3 — AWS by hand: run it on one Linux server (~2 weeks)

**New concept:** the cloud, and a real Linux server.

I do this **by hand** in the AWS console **one time**. In Step 4 I delete it and build it again with Terraform.
Doing it by hand first shows me what Terraform does for me.

⚠️ Do the **AWS cost safety** list first.

- [ ] Install the AWS CLI and log in. `aws sts get-caller-identity` shows my user.
- [ ] Learn the network words: region, availability zone (AZ), VPC, subnet (public / private), internet gateway,
  route table, security group. (The default VPC is OK for this step.)
- [ ] Launch one small EC2 instance (Ubuntu). Security group: port 22 (SSH) and 8080 **only from my IP**.
- [ ] Connect with SSH and a key pair. Why does the key need `chmod 400`?
- [ ] 🐧 On the server: `apt update && apt upgrade`, install Docker, add my user to the `docker` group.
- [ ] 🐧 Look around: `systemctl status docker`, `journalctl -u docker`, `df -h`, `free -h`, `uptime`, `who`.
- [ ] Pull my image from GHCR and start the app with Compose.
- [ ] The app must survive a **reboot**. Reboot the server and check.
- [ ] Run `scripts/smoke-test.sh` against the public IP from my laptop.
- [ ] Delete everything (instance, key pair, security group). Check the bill the next day.

**Experiments:**

- [ ] Remove the 8080 rule from the security group. What does `curl` do now?
  Compare "connection refused" and "timeout". Which one is a firewall, and which one is "nobody listens"?
- [ ] `sudo systemctl stop docker`. What happens to the app? What does `journalctl` show?
- [ ] Stop and start the instance (not reboot). Is the public IP the same? Why?

**Done when I can explain:** region vs AZ; VPC, subnet and security group; public vs private IP;
why SSH is open only to my IP; "connection refused" vs "timeout"; what systemd does.

> **Interview talking point:** "I set up an EC2 server by hand first — network, firewall, SSH, Docker, systemd — so I understood every part before I automated it."

---

## Step 4 — Terraform: infrastructure as code (~2–3 weeks)

**New concept:** Infrastructure as Code (IaC) — the infra is in files, in git, and I can rebuild it with one command.

- [ ] Install Terraform. Learn the loop: `init` → `fmt` → `validate` → `plan` → `apply` → `destroy`.
- [ ] Build Step 3 again, in code: VPC, public subnet, internet gateway, route table, security group, key pair, EC2.
- [ ] A `user_data` script (cloud-init) installs Docker and starts the app when the server boots. No SSH needed.
- [ ] Variables and outputs (for example, print the public IP at the end).
- [ ] **Remote state** in an S3 bucket, with locking. Why must the state not live only on my laptop?
- [ ] Change one thing (for example the instance type). Read the `plan` carefully **before** `apply`.
- [ ] `terraform destroy`, then `terraform apply` again → the app works again. Measure the time.
- [ ] Add `terraform fmt -check` and `terraform validate` to CI.

**Experiments:**

- [ ] Change the security group **by hand** in the console. Then run `terraform plan`. What does it say? (This is called **drift**.)
- [ ] Open the state file. What is inside? Why is it secret?

**Done when I can explain:** declarative vs imperative; what the state is and why it matters;
why we always read `plan` before `apply`; what drift is; why remote state needs locking.

> **Interview talking point:** "All my AWS infra is in Terraform with remote state in S3. I can destroy and rebuild the whole environment with one command in N minutes."

---

## Step 5 — Managed AWS services + automatic deploy (CD) (~2–3 weeks)

**New concept:** managed services, secrets, and Continuous Delivery.

All of this is in **Terraform**, not clicked by hand.

- [ ] **ECR:** a private image registry in AWS. CI pushes images there.
- [ ] **RDS Postgres** instead of the Postgres container. Smallest size, in a **private** subnet.
  Only the app's security group can reach port 5432. Connect with `sslmode=require`.
- [ ] The DB password lives in **SSM Parameter Store** (SecureString). Not in code, not in git, not in Terraform files.
- [ ] The EC2 server gets an **IAM role**, so it can pull images and read the secret. No access keys on the server.
- [ ] App logs go to **CloudWatch Logs**.
- [ ] **CD:** on push to `main`, GitHub Actions deploys the new image.
  It logs in to AWS with **OIDC** — no long-lived AWS keys in GitHub.
- [ ] Deploy flow: build → push to ECR → update the server → run `smoke-test.sh` → if it fails, **roll back** to the last good tag.

**Experiments:**

- [ ] Deploy a broken version (for example `/health` returns 500). Does the pipeline stop it? Does the rollback work?
- [ ] Run `curl` in a loop during a deploy. How many requests fail? (This is the **downtime**.)

**Done when I can explain:** self-managed vs managed (who does backups and updates?);
IAM role vs access keys; what OIDC gives us; why the DB is in a private subnet; how my rollback works.

> **Interview talking point:** "Deploys are automatic: GitHub Actions authenticates to AWS with OIDC, pushes to ECR, deploys, runs a smoke test and rolls back on failure. No static AWS keys anywhere."

---

## Step 6 — Observe and break it (~1.5–2 weeks)

**New concept:** measuring a system, and alerts.

- [ ] 🤖 Prometheus metrics in the service (`/metrics`): request count, error count, latency histogram.
- [ ] 🤖 Structured (JSON) logs. Search them with `jq` locally and with CloudWatch Logs Insights in AWS.
- [ ] Prometheus + Grafana in Compose; build one simple dashboard.
- [ ] A **CloudWatch alarm** (for example: health check fails, or CPU is high) that sends me an email.
- [ ] Write a k6 script for `POST /orders`: ramp up, hold, ramp down.
- [ ] Run it, find the first bottleneck, change one thing, run again, compare.

**Done when I can explain:** p50 vs p95 vs p99 latency and why averages lie; metrics vs logs vs traces;
what a good alert is (it fires on things users feel); what the bottleneck was and how I proved it.

> **Interview talking point:** "I load-tested with k6, saw p99 latency spike at X requests/sec, traced it to Y using Grafana, fixed it, and measured the improvement. An alarm emails me when the health check fails."

---

## Step 7 — Second service and a queue (~2–3 weeks)

**New concept:** service-to-service communication and partial failure — the core distributed systems problem.

- [ ] 🤖 Payment Service in Go with its own database (or own schema). Each service owns its data.
- [ ] First, sync HTTP: Order Service calls Payment Service, with a **timeout**. Add it to Compose, CI and Terraform.
- [ ] Then, async: replace HTTP with an **SQS queue** (in Terraform), plus a **dead-letter queue** (DLQ).
- [ ] Order Service saves the order as `PENDING`, sends a message, and returns `202 Accepted`.
- [ ] Payment Service reads the message, processes the payment, and sends the result back.

**Experiments:**

- [ ] Sync version: stop Payment Service. What happens to new orders?
- [ ] Async version: stop Payment Service, create 10 orders, start it again. What happens? Compare.
- [ ] Make Payment Service crash after the payment but before it deletes the message. Is the order paid twice?
  Fix it: make the consumer **idempotent**.
- [ ] Send a "bad" message that always fails. Where does it go?

**Done when I can explain:** sync vs async trade-offs; why timeouts are mandatory; at-least-once delivery
and why duplicates happen; idempotency; what a DLQ is for; the dual-write problem (and that the outbox pattern fixes it).
SQS and Kafka use the same ideas — I can explain the differences in simple words.

> **Interview talking point:** "Moving to a queue made the system survive payment outages, but brought duplicate messages, so I made the consumer idempotent and added a dead-letter queue."

---

## Step 8 — Kubernetes (~3 weeks)

**New concept:** orchestration — I describe the wanted state, and the system keeps it.

- [ ] Local cluster first (kind or minikube). It is free.
- [ ] Deployments + Services for both services. ConfigMaps and Secrets for config.
- [ ] Readiness and liveness probes (using `/health`). Resource requests and limits.
- [ ] Rolling update, then `kubectl rollout undo`.
- [ ] Scale Payment Service to 3 replicas.
- [ ] Package the app as a **Helm** chart.
- [ ] *Optional:* create EKS with Terraform, deploy, look around for one session, then **destroy it the same day** (it costs money every hour).

**Done when I can explain:** Pod vs Deployment vs Service; desired state vs current state;
why probes matter for zero-downtime deploys; what requests and limits do; Docker Compose vs Kubernetes.

> **Interview talking point:** "I moved the services from Compose to Kubernetes with probes, resource limits and a Helm chart, and practiced rolling updates and rollbacks."

---

## Deliberately left out (for now)

- **Redis** — adds a cache-vs-source-of-truth problem before the basics are solid.
- **Kafka on AWS (MSK)** — expensive. I learn the same queue ideas with SQS.
- **Ansible** — Terraform + cloud-init are enough for one server. Maybe later.
- **GitOps (Argo CD), ECS Fargate** — good next topics after Step 8.
- **Service mesh, multi-region, multi-account** — too much for now.
