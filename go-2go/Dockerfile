FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN apk update &&\
    apk add --no-cache openssl curl gcompat iproute2 &&\
    chmod +x main.go start.sh &&\
    go build -o app main.go

CMD ["./app"]
