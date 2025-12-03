# syntax=docker/dockerfile:1.7
FROM golang:1.25.1-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /out/notification-svc ./services/notifications-service/cmd/notification-svc

FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app
COPY --from=builder /out/notification-svc /app/notification-svc
ENTRYPOINT ["/app/notification-svc"]