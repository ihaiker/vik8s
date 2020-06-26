package asserts

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"strings"
)

func Assert(d *config.Directive, at bool, format string, params ...interface{}) {
	utils.Assert(at, "line %d, %s", d.Line, fmt.Sprintf(format, params...))
}

func ArgsLen(d *config.Directive, size int) {
	Assert(d, len(d.Args) == size, "[%s] args len must is %d: %s", d.Name, size, strings.Join(d.Args, " "))
}

func ArgsMin(d *config.Directive, size int) {
	Assert(d, len(d.Args) >= size, "[%s] args len must is %d: %s", d.Name, size, strings.Join(d.Args, " "))
}

func ArgsRange(d *config.Directive, min, max int) {
	Assert(d, len(d.Args) >= min && len(d.Args) <= max, "[%s] args len must in %d,%d : %s", d.Name, min, max, strings.Join(d.Args, " "))
}
