package main

import (
	"github.com/ihaiker/vik8s/cmd"
	"math/rand"
	"time"
)

var (
	VERSION        = "v0.0.0"
	BUILD_TIME     = "2012-12-12 12:12:12"
	GITLOG_VERSION = "0000"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	cmd.Execute(VERSION, BUILD_TIME, GITLOG_VERSION)
}
