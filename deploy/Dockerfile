FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp ./cmd/api

FROM alpine:latest
RUN apk add curl
WORKDIR /root/
COPY --from=builder /app/myapp .
COPY --from=builder /app/docs ./docs
CMD ["./myapp"]