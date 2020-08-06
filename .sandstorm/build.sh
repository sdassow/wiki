#!/bin/bash
set -euo pipefail

export GOPATH=$HOME/go

[ -d $GOPATH/src/github.com/prologic/ ] || mkdir -p $GOPATH/src/github.com/prologic/
[ -L $GOPATH/src/github.com/prologic/wiki ] || \
	ln -s /opt/app $GOPATH/src/github.com/prologic/wiki

cd /opt/app
# cd $GOPATH/src/github.com/prologic/wiki
go get -v -d ./...
go build .
exit 0
