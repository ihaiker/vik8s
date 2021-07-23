package paths

import (
	"os"
	"path/filepath"
)

type Json map[string]interface{}

var ConfigDir = os.ExpandEnv("$HOME/.vik8s")
var Cloud = "default"
var China = true

func Join(path ...string) string {
	paths := append([]string{os.ExpandEnv(ConfigDir), Cloud}, path...)
	outpath := filepath.Join(paths...)
	outpath, _ = filepath.Abs(outpath)
	return outpath
}

func Vik8sConfiguration() string {
	return Join("vik8s.conf")
}

func HostsConfiguration() string {
	return Join("hosts.conf")
}
