FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy

RUN go build -o darkchat .

FROM golang:1.21

WORKDIR /app
COPY --from=builder /app /app

ENTRYPOINT ["go", "test", "-v", "./..."]
