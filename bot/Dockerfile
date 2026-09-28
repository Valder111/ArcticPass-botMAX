FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/arcticpass-bot ./cmd/main

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 65532 -s /sbin/nologin app

WORKDIR /app

COPY --from=builder /out/arcticpass-bot /app/arcticpass-bot

USER app

ENTRYPOINT ["/app/arcticpass-bot"]
