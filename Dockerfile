FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o image-processor cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y libc6 ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/image-processor /app/

EXPOSE 8080

CMD ["/app/image-processor"]
