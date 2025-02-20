FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy


RUN go build -o darkchat .


FROM golang:1.21

WORKDIR /app
COPY --from=builder /app /app

ENTRYPOINT ["go", "test", "-v", "./..."]
