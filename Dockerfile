FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p bin \
 && CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o bin/main ./cmd/main.go

FROM alpine:3.21 AS runtime

RUN adduser -D appuser \
 && rm -rf /var/cache/apk/*

# copy while still root
COPY --from=builder /app/bin/main /usr/local/bin/main

USER appuser
WORKDIR /home/appuser

EXPOSE 8080
CMD ["/usr/local/bin/main"]


