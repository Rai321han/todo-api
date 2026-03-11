# Todo API

A RESTful Todo API built with Go, Beego, PostgreSQL, and JWT authentication.

## Content Outline

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [API Base URL and Versioning](#api-base-url-and-versioning)
- [Prerequisites](#prerequisites)
- [Installation and Setup](#installation-and-setup)
- [Run the Application](#run-the-application)
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

- Go (`go.mod` currently targets `1.25`)
- Beego v2
- PostgreSQL 16
- JWT (`github.com/golang-jwt/jwt/v5`)
- Docker Compose (for local database)

## Project Structure

```text
todo-api/
├── conf/
│   ├── app.conf
│   └── app.sample.conf
├── controllers/
│   ├── auth.go
│   └── todo.go
├── init-db/
│   └── init.sql
├── middlewares/
│   ├── auth_middleware.go
│   ├── global_exception.go
│   └── request_logger.go
├── models/
│   ├── db/
│   │   └── db.go
│   ├── todo/
│   │   ├── repository.go
│   │   └── todo.go
│   └── user/
│       ├── repository.go
│       └── user.go
├── routers/
│   └── router.go
├── service/
│   ├── auth/
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── todo/
│       ├── todo.go
│       └── todo_test.go
├── utils/
│   └── api_error.go
├── docker-compose.yml
├── go.mod
├── go.sum
└── main.go
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
cp conf/app.sample.conf conf/app.conf
```

Then update `conf/app.conf` with your database and JWT secret values

4. Start PostgreSQL using Docker Compose.

```bash
docker compose up -d
```

This also runs schema initialization from `init-db/init.sql`.

5. Download Go dependencies.

```bash
go mod tidy
```

## Run the Application

Start the API server:

```bash
bee run
```

If successful, the service will be available at:

```text
http://localhost:8080
```

## API Endpoints

Auth:

- `POST /v1/api/auth/register`
- `POST /v1/api/auth/login`

Todos (JWT required):

- `POST /v1/api/todos/`
- `GET /v1/api/todos/`
- `GET /v1/api/todos/:id`
- `PUT /v1/api/todos/:id`
- `DELETE /v1/api/todos/:id`

Todo list query parameters (`GET /v1/api/todos/`):

- `status`: `completed` or `pending`
- `sort_by`: `created_at` or `title`
- `order`: `asc` or `desc`
- `search`: search text on title
- `page`: integer, minimum `1`
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

5. Error format

Most validation and business errors are returned as:

```json
{
  "error": "message",
  "code": 400
}
```

## Postman Documentation

Official Postman docs for this API:

- https://documenter.getpostman.com/view/45965098/2sBXierZWC#4b99640c-b16c-43d9-ab03-62d9fe455d10

## Run Tests

```bash
go test ./...
```