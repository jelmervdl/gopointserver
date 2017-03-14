#!/bin/sh
env CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .
mv ./gopointserver ./gopointserver-amd64-static