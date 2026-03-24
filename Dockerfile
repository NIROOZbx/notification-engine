FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/backend ./cmd/

FROM alpine:3.23

WORKDIR /app

RUN adduser -D nirooz

USER nirooz

COPY --from=builder /app/backend .

COPY config/config.yaml config/config.yaml

EXPOSE 8080

CMD [ "./backend" ]
