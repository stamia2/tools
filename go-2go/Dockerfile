FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git musl-dev gcc openssl
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app main.go

FROM alpine:3.18
RUN apk add --no-cache gcompat openssl
COPY --from=builder /app/app /app/
WORKDIR /app
CMD ["./app"]
