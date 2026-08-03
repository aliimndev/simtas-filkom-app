# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o simtas-api ./cmd/server
# Job 29: production seed command (idempotent admin/kaprodi account seeding)
RUN go build -o simtas-seed ./cmd/seed

# Stage 2: Run
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/simtas-api .
COPY --from=builder /app/simtas-seed .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./simtas-api"]
