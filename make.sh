#!/usr/bin/env sh

VERSION=1.0.0

echo fetch dependencies
go get github.com/tarm/serial
go get github.com/Sirupsen/logrus
go get github.com/yosssi/gmq/mqtt
go get github.com/yosssi/gmq/mqtt/client

echo build linux/arm/5
mkdir -p release/$VERSION/linux/arm
GOOS=linux GOARCH=arm GOARM=5 go build mqtt-teleinfo.go
mv mqtt-teleinfo release/$VERSION/linux/arm/mqtt-teleinfo

echo build linux/amd64
mkdir -p release/$VERSION/linux/amd64
GOOS=linux GOARCH=amd64 go build mqtt-teleinfo.go
mv mqtt-teleinfo release/$VERSION/linux/amd64/mqtt-teleinfo
