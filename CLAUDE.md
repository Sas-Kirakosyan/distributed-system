# Project context

## About me

- I am a JavaScript software engineer. I want a job as a Cloud engineer or DevOps engineer.
- I already know general programming, JSON, HTTP, REST APIs, status codes and basic Git. Do not explain these.
- I am here to learn Linux, Docker, CI/CD, AWS, Terraform, monitoring and Kubernetes. Explain these from the root concept.
- If a comparison with JavaScript makes something faster to understand, use it. For example: `jq -r .id` is like `JSON.parse(text).id`.

## How we work

- The plan is in `README.md`. Tickets are in `tasks/STEP-N.md`.
- Claude writes the Go code and Go tests only.
- I do everything else from a task and hints: Docker, Compose, SQL, bash, CI, AWS, Terraform and config. Do not give me the full answer.
- Claude reviews what I push or paste, and says what is wrong and why.
- Use short, simple B1-level English. Teach the concept first and the tool name second.

## My computer

- Windows with WSL2. I work in the Ubuntu terminal, in `/mnt/c/Users/saski/OneDrive/Desktop/destrubuted-system`.
- Windows has its own Postgres on ports 5432 and 5433. Do not publish Docker ports on them.
- Windows reserves TCP ports 50000–50059, 55044–55443 and 55508–55607 (checked 2026-11-18; the ranges can change after a reboot). Docker cannot publish ports there. Use for example 15432 for a temporary Postgres.
