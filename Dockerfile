FROM golang:1.23.6 AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:3.18 AS final

RUN apk update && \
    apk add --no-cache sed ca-certificates && \
    rm -rf /var/cache/apk/*
RUN apk update && apk add --no-cache jq
WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations
CMD ["./app"]
