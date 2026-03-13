FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work* ./

COPY pkg/ ./pkg/

COPY services/backend/go.mod services/backend/go.sum ./services/backend/

WORKDIR /app/services/backend

RUN go mod download

WORKDIR /app

COPY services/backend/ ./services/backend/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/backend ./services/backend/cmd/

FROM alpine:3.23

WORKDIR /app

RUN adduser -D nirooz

USER nirooz

COPY --from=builder /app/backend .

COPY services/backend/config/config.yaml ./config/config.yaml

EXPOSE 8080

CMD [ "./backend" ]
