package tools

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
	return filepath.Join(paths...)
}
