# ----------- Stage 1: Build -----------
FROM golang:1.24.4-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server main.go

# ----------- Stage 2: Run ------------
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./server"]
