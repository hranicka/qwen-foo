FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o hello ./cmd/server/
RUN go build -o migrate ./cmd/migrate/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/hello .
COPY --from=builder /app/migrate .
COPY config.yaml .
EXPOSE 8080
ENTRYPOINT ["sh", "-c", "./migrate up && ./hello"]
