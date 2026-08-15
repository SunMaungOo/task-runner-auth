FROM golang:alpine3.24 AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o task-runner-auth ./cmd

FROM alpine:3.23.5 AS runner

RUN apk --no-cache add ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 10001

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/task-runner-auth .

USER appuser

CMD ["./task-runner-auth"]