#!/bin/bash

Version=$(git describe --tags `git rev-list --tags --max-count=1`)
BuildDate=$(date +"%F %T")
GitCommit=$(git rev-parse HEAD)
param="-X main.VERSION=${Version} -X main.GITLOG_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'"

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

if [ "$1" == "all" ]; then
  echo "build full platform package"
else
  echo "build vik8s"
  go build -trimpath -ldflags "$param" -o ./bin/vik8s main.go
fi
