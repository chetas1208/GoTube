FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/api /app/api
COPY backend/worker /app/worker

WORKDIR /app/worker
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.20

RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=builder /worker /worker

RUN mkdir -p /tmp/gotube-worker

EXPOSE 9091

CMD ["/worker"]
