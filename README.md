```markdown
# Personal Task Tracker API

A lightweight, robust, and containerized RESTful API for managing a personal list of tasks. Built from scratch using Go, the Gin framework, and a CGO-free SQLite driver. Designed to run seamlessly in isolated environments (such as WSL/Docker) without external dependencies.

## Architecture & Tech Stack

- **Language:** Go (1.26+)
- **Web Framework:** [Gin Gonic](https://github.com/gin-gonic/gin) (chosen for its high-performance routing, built-in logging, and clean middleware ecosystem)
- **Database:** SQLite via `github.com/glebarez/go-sqlite`
- **Containerization:** Docker & Docker Compose (Multi-stage production build)

### Engineering Decisions & Justifications

1. **Pure Go SQLite Driver (CGO-free):**
   Instead of using the standard `mattn/go-sqlite3` driver which relies on GCC and CGO, this project implements `github.com/glebarez/go-sqlite`. This architectural choice completely eliminates dynamic linking issues during cross-compilation, avoids heavy Windows/WSL security blocking policies (like AppLocker blocking host `.exe` execution), and guarantees highly deterministic builds inside minimal Scratch/Alpine Docker stages.

2. **Pointers for Partial Updates (PATCH `/tasks/:id`):**
   To distinguish between a field explicitly sent as an empty string `""` versus a field completely omitted from the JSON request payload (`nil`), the request structure utilizes Go pointers (`*string`). This ensures that if a user updates only the `status`, the existing `title` in the database remains completely untouched.

3. **Domain & Business Logic Validation Integration:**
   A critical business constraint states: *A task cannot be marked as `done` (or created as `done`) if it has an empty title*. This is enforced deterministically by pulling the current task state from the database, performing a memory-level structural merge with the user's input, and executing validation immediately before updating the storage layers.

4. **Handling Ambiguity:**
   The assignment specification left a minor ambiguity regarding the exact storage format and constraints of the `due_date` field. We made a conscious decision to handle it as an ISO 8601 string (`YYYY-MM-DD`). This provides flexibility for timezone-agnostic frontend rendering while keeping database serialization simple and performant.

---

## Getting Started

### Prerequisites
- Docker and Docker Compose installed within your environment (fully compatible with WSL/Ubuntu).

### Running the Application

Forget about long Docker run commands with complex volume passing. The deployment is completely declarative. Run the following command in the project root directory:

```bash
docker compose up --build -d

```

This command will:

1. Trigger a multi-stage Docker build that safely compiles the Go binary inside Linux.
2. Spin up the API server bound to host port `8080`.
3. Auto-initialize an internal Docker volume (`task_tracker_data`) to ensure absolute data persistence across restarts.
4. Auto-create the SQLite database file (`tasks.db`) and its schema if they do not exist.

### Verifying the Deployment

To monitor application logs and verify successful initialization, run:

```bash
docker compose logs -f task-tracker-api

```

To shut down the service completely:

```bash
docker compose down

```

### Running Integration Tests

Automated integration tests run within an isolated in-memory environment via a temporary container to keep the host database clean. To execute the test suite, run:

```bash
docker run --rm -v $(pwd):/app -w /app golang:alpine go test -v ./...

```

---

## API Endpoint Specification

### 1. Create a New Task

* **HTTP Method:** `POST`
* **Path:** `/tasks`
* **Headers:** `Content-Type: application/json`
* **Payload:**

```json
{
  "title": "Implement Docker Compose Configuration",
  "due_date": "2026-07-05"
}

```

* **Success Response (`201 Created`):**

```json
{
  "id": 1,
  "title": "Implement Docker Compose Configuration",
  "status": "pending",
  "due_date": "2026-07-05",
  "created_at": "2026-07-03T19:40:00Z"
}

```

### 2. List All Tasks (With Filtering & Pagination)

* **HTTP Method:** `GET`
* **Path:** `/tasks`
* **Query Parameters:** - `status` (Optional. Values: `pending` or `done`)
* `page` (Optional. Default: `1`)
* `limit` (Optional. Default: `10`)


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
* **Graceful DB Connection Shutdown:** Implemented connection pools cleanup mechanisms via `defer db.Close()` on application signals.

```

```
