package main

import (
	"github.com/ihaiker/vik8s/cmd"
	"math/rand"
	"time"
)

var (
	version = "v0.0.0"
	date    = "2012-12-12 12:12:12"
	commit  = "0000"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	cmd.Execute(version, date, commit)
}
