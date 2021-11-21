#!/bin/bash

export BASE_PATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )"
cd $BASE_PATH/..

Version=$(git describe --tags `git rev-list --tags --max-count=1`)
BuildDate=$(date +"%F %T")
GitCommit=$(git rev-parse HEAD)
param="-X main.version=${Version} -X main.commit=${GitCommit} -X 'main.date=${BuildDate}'"

go_bindata_bin=$(which go-bindata)
if [ "$go_bindata_bin" == "" ]; then
  echo "install go-bindata"
  go get -u github.com/shuLhan/go-bindata/cmd/go-bindata
else
  echo "go-bindata installed $go_bindata_bin"
fi

echo "generator go bin data"
go-bindata -modtime 1590460659 -pkg yamls -o yaml/assets.go -ignore .*\.go -ignore .*\.part yaml/...

echo "format yaml/assets.go"
go fmt yaml/assets.go

go build -trimpath -ldflags "$param" -o ./bin/vik8s cmd/main.go
