FROM golang:1.24.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base

COPY . .

ENV CGO_ENABLED=1

RUN go build -o bintracebench.out ./cmd/bintracebench
RUN go build -o bintracer.out ./internal/tools/bintracer.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bintracebench.out .
COPY --from=builder /app/bintracer.out .


EXPOSE 8080

ENV DB_TYPE=sqlite
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080

ENTRYPOINT ["./bintracebench.out"]
