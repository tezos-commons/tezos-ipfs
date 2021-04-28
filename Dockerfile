# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /go/src/github.com/tezos-commons/tezos-ipfs/
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/tezos-commons/tezos-ipfs/app .
ENTRYPOINT ["./app"]
