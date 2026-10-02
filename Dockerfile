# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache module downloads in their own layer, invalidated only when
# go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates: outbound HTTPS (Termii, SMTP-over-TLS). tzdata: so
# time.LoadLocation and friends work if ever needed; the app itself
# always operates in UTC internally.
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api .

USER app
EXPOSE 8080

ENTRYPOINT ["./api"]
