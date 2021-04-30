#!/bin/bash


CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app main.go
docker build -f Dockerfile.quickbuild -t tipfs:test .
rm app
cd test
docker-compose rm -f
docker-compose up