# Personal Task Tracker API

A lightweight, robust, and containerized RESTful API for managing a personal list of tasks. Built from scratch using Go, the Gin framework, and a CGO-free SQLite driver. Designed to run seamlessly in isolated environments (such as WSL/Docker) without external dependencies.

## Architecture & Tech Stack

* **Language:** Go (1.26+)
* **Web Framework:** [Gin Gonic](https://github.com/gin-gonic/gin) (chosen for its high-performance routing, built-in logging, and clean middleware ecosystem)
* **Database:** SQLite via `github.com/glebarez/go-sqlite`
* **Containerization:** Docker & Docker Compose (Multi-stage production build)

### Engineering Decisions & Justifications

1. **Pure Go SQLite Driver (CGO-free):**
Instead of using the standard `mattn/go-sqlite3` driver which relies on GCC and CGO, this project implements `github.com/glebarez/go-sqlite`. This architectural choice completely eliminates dynamic linking issues during cross-compilation, avoids heavy Windows/WSL security blocking policies (like AppLocker blocking host `.exe` execution), and guarantees highly deterministic builds inside minimal Scratch/Alpine Docker stages.
2. **Pointers for Partial Updates (PATCH `/tasks/:id`):**
To distinguish between a field explicitly sent as an empty string `""` versus a field completely omitted from the JSON request payload (`nil`), the request structure utilizes Go pointers (`*string`). This ensures that if a user updates only the status, the existing title remains untouched instead of being wiped out by empty defaults.

---

## Quick Start (How to Run)

To spin up the application infrastructure locally for the first time, execute:

```bash
# Build the multi-stage image and run the stack in detached mode
docker compose up --build -d

```

The API engine will initialize the SQLite schema and start listening for inbound traffic on `http://localhost:8080`.

---

## Infrastructure Management (Isolated Rebuilds)

To completely purge application state, clear persistent volumes, and force a clean un-cached binary rebuild inside WSL/Docker, execute:

```bash
docker compose down -v && docker compose build --no-cache && docker compose up -d

```

---

## API Endpoints & Usage Examples

### 1. Create a Task

* **HTTP Method:** `POST`
* **Path:** `/tasks`
* **Payload:**

```json
{
  "title": "Buy groceries",
  "status": "pending",
  "due_date": "2026-05-15"
}

```

* **Success Response (`201 Created`)**

### 2. List Tasks (With Pagination & Filtering)

* **HTTP Method:** `GET`
* **Path:** `/tasks`
* **Query Parameters:**
* `status` (Optional: `pending` or `done`)
* `page` (Optional, default: `1`)
* `limit` (Optional, default: `10`)


* **Sorting:** Strictly auto-sorted by creation date descending (`created_at DESC`).
* **Examples:**
* Get all tasks (Default page/limit): `GET http://localhost:8080/tasks`
* Filter completed tasks with custom pagination: `GET http://localhost:8080/tasks?status=done&page=2&limit=5`



### 3. Get Single Task By ID

* **HTTP Method:** `GET`
* **Path:** `/tasks/:id`
* **Success Response (`200 OK`):** Returns the specific JSON task object.
* **Error Response (`404 Not Found`):** `{"error": "Task not found"}`

### 4. Partially Update a Task

* **HTTP Method:** `PATCH`
* **Path:** `/tasks/:id`
* **Payload:** (Can pass any combination of fields)

```json
{
  "status": "done"
}

```

* **Error Response (`400 Bad Request`):** Sent if an invalid status value is passed or if the update results in a `done` task with an empty title.

### 5. Delete a Task

* **HTTP Method:** `DELETE`
* **Path:** `/tasks/:id`
* **Success Response (`200 OK`):** `{"message": "Task deleted successfully"}`

---

## Production Security & Safety Edge-Cases Handled

* **Strict Title Length Constraints:** Enforced a `max 200` character limit to protect the data layer against buffer-bloat payloads.
* **SQL Injection Mitigation:** Every interaction with the SQLite database utilizes parameterized query placeholders (`?`) rather than unchecked string formatting.

---

## Retrospective & Future Improvements

1. **Database Connection Pooling & Tuning:** Currently, the global `sql.DB` instance handles open channels implicitly. If I had more time, I would explicitly configure `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` to prevent SQLite lock contention under high concurrent write loads.
2. **Structured Service Layer Architecture:** The business and validation logic is currently coupled directly inside the controller execution layer (`main.go`). Extracting an isolated Domain Service layer would significantly improve testability and separate transport logic from core rules.
3. **Comprehensive Integration Test Suite:** While validation constraints are verified using an in-memory SQLite state, expanding the testing infrastructure to completely cover paginated matrix queries and partial updates validation edge-cases would increase long-term reliability.

## Assumptions Made

* **Timezone Standard:** The specification left storage representations ambiguous. I assumed standardizing metadata serialization onto ISO 8601 strings (`YYYY-MM-DD` and RFC3339 timestamps) satisfies both deterministic sorting (`created_at DESC`) and client-side timezone rendering neutrality.
