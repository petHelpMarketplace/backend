FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# make the bin/ dir
RUN mkdir -p bin

# now run go build as its own step (no fancy quoting)
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/main ./cmd/pethelp/main.go

FROM alpine:3.21 AS runtime
RUN adduser -D appuser \
 && rm -rf /var/cache/apk/*

COPY --from=builder /app/bin/main /usr/local/bin/main

USER appuser
WORKDIR /home/appuser

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/main"]


