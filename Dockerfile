FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

COPY . .

RUN go build -o todo-app main.go

# Final lightweight image
FROM alpine:3.14
WORKDIR /app

# Copy built binary
COPY --from=builder /app/todo-app .
COPY conf/ ./conf

# Expose port
EXPOSE 8080

# Command to run
CMD ["./todo-app"]