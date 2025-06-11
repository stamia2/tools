FROM golang:1.20-alpine AS builder

RUN apk add --no-cache git gcc musl-dev openssl curl \
    iproute2 gcompat make

WORKDIR /app



COPY . .

RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -o app \
    main.go

FROM alpine:3.18
RUN apk add --no-cache \
    ca-certificates \
    openssl \
    curl \
    gcompat \
    iproute2

RUN mkdir -p /app/tmp
WORKDIR /app
COPY --from=builder /app/app /app/

EXPOSE 3000
CMD ["/app/app"]
