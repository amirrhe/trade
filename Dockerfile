# syntax=docker/dockerfile:1

FROM golang:1.22.4-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/trade ./cmd/http

FROM alpine:3.18 AS runtime

RUN apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser
USER appuser

COPY --from=builder /app/trade .

COPY --from=builder /src/docs /home/appuser/docs

EXPOSE 8080

ENTRYPOINT ["./trade"]
