FROM golang:1.8-alpine
RUN apk update
RUN apk add openssl ca-certificates git
RUN mkdir -p /go/src/oct-redis-api
ADD server.go  /go/src/oct-redis-api/server.go
ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh
CMD ["/go/src/oct-redis-api/server"]
EXPOSE 3000

