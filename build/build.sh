#!/bin/bash

appname=template

time=`date +%F` 
gover=`go version`
githash=`git rev-list HEAD -n 1 | cut -c 1-`

cd main

go build -mod=vendor -i -v -o ../bin/${appname} -ldflags  "-s -w -X 'main.BUILDTIME=${time}' -X 'main.GOVERSION=${gover}' -X 'main.GITHASH=${githash}'"