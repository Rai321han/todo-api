# Todo API

A RESTful Todo API built with Go, Beego, PostgreSQL, and JWT authentication.

## Content Outline

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [API Base URL and Versioning](#api-base-url-and-versioning)
- [Prerequisites](#prerequisites)
- [Installation and Setup](#installation-and-setup)
- [API Endpoints](#api-endpoints)
- [Request and Response Examples](#request-and-response-examples)
- [Postman Documentation](#postman-documentation)
- [Run Tests](#run-tests)

## Overview

This project provides:

- User registration and login
- JWT-based authentication for protected routes
- CRUD operations for todo items
- Todo listing with filtering, search, sorting, and pagination

## Tech Stack

- Go 1.25
- Beego v2
- PostgreSQL 16
- JWT (`github.com/golang-jwt/jwt/v5`)
- GoConvey for testing
- Docker

## Project Structure

```text
todo-api/                              # Project root
├── conf/                              # Application configuration files
│
├── controllers/                       # HTTP handlers (API endpoints)
│
├── init-db/                           # Database bootstrap scripts
│   └── init.sql                       # Initial schema and setup SQL
│
├── middlewares/                       # Request/response middleware layer
│   ├── auth_middleware.go             # JWT authentication middleware
│   ├── global_exception.go            # Global error recovery/handling
│   └── request_logger.go              # HTTP request logging middleware
│
├── models/                            # Data models and DB repositories
│   ├── db/                            # DB connection setup
│   ├── todo/                          # Todo model and repository logic
│   └── user/                          # User model and repository logic
│
├── routers/                           # API route registration
│   └── router.go                      # Route definitions and version groups
│
├── service/                           # Business logic layer
│   ├── auth/                          # Auth business logic and tests
│   └── todo/                          # Todo business logic and tests
│
├── utils/                             # Shared utility helpers
│
├── docker-compose-sample.yml          # Sample docker compose for project build and run
├── go.mod                             # Go module dependencies
├── go.sum                             # Dependency checksum lock file
└── main.go                            # Application entrypoint
```

## API Base URL and Versioning

Base URL (local):

```text
http://localhost:8080
```

API prefix:

```text
/v1/api
```

## Prerequisites

Install the following:

- Go (recommended: `1.25+` to match `go.mod`)
- Docker and Docker Compose
- Git
- `curl` (optional, for testing)

## Installation and Setup

1. Clone the repository.

```bash
git clone https://github.com/Rai321han/todo-api.git
cd todo-api
```

2. Configure application settings.

```bash
mv conf/app.sample.conf conf/app.conf
```

Then update `conf/app.conf` with your database and JWT secret values

3. Update compose file
```bash
mv docker-compose-sample.yml docker-compose.yml
```

Then update the variables.

4. Build the app and run servers

```bash
docker compose up --build
```

This also runs schema initialization from `init-db/init.sql`.


If successful, the api service will be available at:

```text
http://localhost:<defined_port>
```

## API Endpoints

Auth:

- Register user: `POST /v1/api/auth/register`
- Login user: `POST /v1/api/auth/login`

Todos (JWT Token required):

- Create Todo: `POST /v1/api/todos/`
- Get All Todo: `GET /v1/api/todos/`
- Get todo by id: `GET /v1/api/todos/:id`
- Update by id: `PUT /v1/api/todos/:id`
- Delete by id: `DELETE /v1/api/todos/:id`

Todo list query parameters (`GET /v1/api/todos/`):

- `status`: `completed` or `pending`
- `sort_by`: `created_at` or `title`
- `order`: `asc` or `desc`
- `search`: search text on title
- `page`: integer,
- `limit`: integer,

## Request and Response Examples

1. Register user

```bash
curl -X POST http://localhost:8080/v1/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "password123"
  }'
```

Success response (`201`):

```json
{
  "message": "user created"
}
```

2. Login

```bash
curl -X POST http://localhost:8080/v1/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

Success response (`200`):

```json
{
  "token": "<jwt-token>"
}
```

3. Create todo (authorized)

```bash
curl -X POST http://localhost:8080/v1/api/todos/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "title": "Finish internship task",
    "description": "Write README and verify endpoints",
    "is_completed": false
  }'
```

Success response (`201`):

```json
{
  "id": 1,
  "title": "Finish internship task",
  "user_id": 1,
  "description": "Write README and verify endpoints",
  "is_completed": false,
  "created_at": "2026-03-11T10:00:00Z",
  "updated_at": "2026-03-11T10:00:00Z"
}
```

4. List todos with filters

```bash
curl "http://localhost:8080/v1/api/todos/?status=pending&sort_by=created_at&order=desc&page=1&limit=10&search=internship" \
  -H "Authorization: Bearer <jwt-token>"
```

Response shape (`200`):

```json
{
  "total_pages": 1,
  "current_page": 1,
  "limit": 10,
  "total_count": 1,
  "todos": [
    {
      "id": 1,
      "title": "Finish internship task",
      "user_id": 1,
      "description": "Write README and verify endpoints",
      "is_completed": false,
      "created_at": "2026-03-11T10:00:00Z",
      "updated_at": "2026-03-11T10:00:00Z"
    }
  ]
}
```

5. Get todo by id

```bash
curl -X GET http://localhost:8080/v1/api/todos/1 \
  -H "Authorization: Bearer <jwt-token>"
```

Success response (`200`):

```json
{
  "id": 1,
  "title": "Finish internship task",
  "user_id": 1,
  "description": "Write README and verify endpoints",
  "is_completed": false,
  "created_at": "2026-03-11T10:00:00Z",
  "updated_at": "2026-03-11T10:00:00Z"
}
```

6. Update todo by id

```bash
curl -X PUT http://localhost:8080/v1/api/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "title": "Finish internship task (updated)",
    "description": "Update endpoint example in README",
    "is_completed": true
  }'
```

Success response (`200`):

```json
{
  "id": 1,
  "title": "Finish internship task (updated)",
  "user_id": 1,
  "description": "Update endpoint example in README",
  "is_completed": true,
  "created_at": "2026-03-11T10:00:00Z",
  "updated_at": "2026-03-12T09:30:00Z"
}
```

7. Delete todo by id

```bash
curl -X DELETE http://localhost:8080/v1/api/todos/1 \
  -H "Authorization: Bearer <jwt-token>"
```

Success response (`204`):

No response body.

8. Get all todos (simple)

```bash
curl -X GET http://localhost:8080/v1/api/todos/ \
  -H "Authorization: Bearer <jwt-token>"
```

Response shape (`200`):

```json
{
  "total_pages": 1,
  "current_page": 1,
  "limit": 10,
  "total_count": 1,
  "todos": [
    {
      "id": 1,
      "title": "Finish internship task",
      "user_id": 1,
      "description": "Write README and verify endpoints",
      "is_completed": false,
      "created_at": "2026-03-11T10:00:00Z",
      "updated_at": "2026-03-11T10:00:00Z"
    }
  ]
}
```

9. Error format

Most validation and business errors are returned as:

```json
{
  "error": "message",
  "code": 400
}
```

## Postman Documentation

Official Postman docs for this API:

- [API DOCUMENTATION](https://documenter.getpostman.com/view/45965098/2sBXierZWC#4b99640c-b16c-43d9-ab03-62d9fe455d10)

## Run Tests

```bash
go test ./...

or

$GOPATH/bin/goconvey
```