# DVAPI

**Damn Vulnerable API (DVAPI)** is an intentionally vulnerable API application designed for learning, practicing, and testing API security concepts.

The project provides a realistic development environment where developers and security researchers can explore common API vulnerabilities in a safe and controlled setup.

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

---

## Quick start

1. Clone the repo and enter it:

```bash
git clone <repo-url> DVAPI
cd DVAPI
```

2. Start the stack (build images, migrate the database, and load sample data):

```bash
make up
```

That starts everything, sets up the database, and loads sample data.

3. Open the **App (UI)** and create an account with **Register**, or log in if you already have one.

Stop everything with:

```bash
make down
```

To wipe database and cache volumes as well:

```bash
make clean
```
---
## What’s running

| Component           | Purpose                                                   |                           Local Port |
| ------------------- | --------------------------------------------------------- | -----------------------------------: |
| **Frontend**        | Vue.js web application                                    |                             **5173** |
| **Backend**         | Go-based REST API                                         |                             **8080** |
| **Library Service** | Flask helper                                              |                             **5000** |
| **PostgreSQL**      | Primary relational database                               |                             **5433** |
| **Redis**           | In-memory cache and session store                         |                             **6380** |
| **MailHog**         | Local SMTP server and email testing UI                    |                             **1025** |
| **Loki**            | Centralized log aggregation                               |                             **3100** |
| **Grafana**         | Monitoring and log visualization (anonymous admin access) |                             **3001** |

---

## Make commands

```bash
make up          # start everything (build if needed), migrate, seed
make down        # stop and remove containers
make ps          # what’s running?
make logs        # follow all logs
make clean       # stop and wipe database/cache data
```

`make dev` does the same thing as `make up`.

### Database

```bash
make migrate-up       # apply new schema changes
make migrate-down     # undo the last change
make migrate-status   # see migration state
```

### Sample data

The stack must already be up (`make up`).

```bash
make seed                           # load all default sample data
make seed ARGS="users"              # only users
make seed ARGS="courses enrollments"
make seed-status                    # what’s already loaded?
```

Default seed sets: users, courses, enrollments, assignments, submissions, announcements.

### Other useful bits

```bash
make rebuild          # rebuild images from scratch
make prune            # remove containers, data volumes, and local images
make backend-shell    # shell inside the backend container
make db-shell         # open a database prompt
```

Run `make help` for the full list.

---

## Vulnerability surface

DVAPI intentionally includes vulnerabilities across multiple attack surfaces and protocols to simulate realistic API security challenges.

### Surfaces

- REST API
- GraphQL
- SMTP
- Database layer
- Operating system (OS) layer

### Categories

- OWASP API Security Top 10
- Injection vulnerabilities (across multiple protocols and technologies)
- Authentication and authorization flaws
- Information disclosure
- Input validation failures

> **Note:** Specific vulnerabilities and exploitation paths are intentionally undocumented. Discovering and exploiting them is part of the learning experience.
