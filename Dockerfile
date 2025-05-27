FROM golang:1.21.7-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/main cmd/main.go

FROM alpine:3.21 AS runtime

RUN adduser -D appuser && rm -rf /var/cache/apk/*
USER appuser

WORKDIR /home/appuser
COPY --from=builder /app/bin/main /usr/local/bin/

EXPOSE 8080

CMD ["/usr/local/bin/main"]

