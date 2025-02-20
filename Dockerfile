FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/darkchat /app/darkchat

RUN chmod +x /app/darkchat

EXPOSE 8080

ENTRYPOINT [ "./darkchat run" ]