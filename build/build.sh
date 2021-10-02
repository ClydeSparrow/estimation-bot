#!/bin/sh
set -eux
gofmt -s -w .
reset
go build -o estimation .