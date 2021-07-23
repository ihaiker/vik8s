package plugins

import (
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"plugin"
)

type (
	ReduceHandler struct {
		Names   []string
		Demo    string
		Handler func(version, prefix string, item *config.Directive) (obj metav1.Object)
	}

	PluginLoad func() ReduceHandler

	ReduceHandlers []ReduceHandler
)

var Manager = ReduceHandlers{}

func (m *ReduceHandler) Has(name string) bool {
	for _, n := range m.Names {
		if n == name {
			return true
		}
	}
	return false
}

func (m *ReduceHandlers) Handler(version, prefix string, item *config.Directive) (metav1.Object, bool) {
	name, _ := utils.Split2(item.Name, ":")
	for _, reduce := range *m {
		if reduce.Has(name) {
			obj := reduce.Handler(version, prefix, item)
			return obj, true
		}
	}
	return nil, false
}

func plugins() []string {
	files := make([]string, 0)
	pluginDir := paths.Join("plugins", "reduce")
	filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func Load() {
	pluginFiles := plugins()
	for _, file := range pluginFiles {
		p, err := plugin.Open(file)
		utils.Panic(err, "load reduce plugins %s", file)
		sl, err := p.Lookup("Reduce")
		utils.Assert(err == nil, "load reduce plugins %s", file)
		Manager = append(Manager, sl.(PluginLoad)())
	}
}
