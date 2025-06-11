FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git musl-dev gcc openssl
WORKDIR /app

COPY . .

RUN if [ ! -f go.mod ]; then go mod init your-module-name; fi

RUN go mod tidy && go mod download

RUN  go build -o app main.go

FROM alpine:3.18
RUN apk add --no-cache gcompat openssl
COPY --from=builder /app/app /app/
CMD ["./app"]
