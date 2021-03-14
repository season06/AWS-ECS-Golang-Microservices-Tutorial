FROM golang:latest

WORKDIR /web
COPY . /web

RUN go mod init web
RUN go get github.com/go-redis/redis
RUN go build

EXPOSE 8000
ENTRYPOINT ./web