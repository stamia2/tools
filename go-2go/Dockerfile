FROM golang:1.20-alpine AS builder

RUN apk add --no-cache git musl-dev gcc openssl curl make

WORKDIR /app

RUN ls -la

RUN test -f go.mod || go mod init your-module-name

RUN go mod download -x

RUN go list -m all

RUN CGO_ENABLED=0 go build -gcflags="all=-N -l" -o app main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/app /app/
CMD ["/app/app"]
