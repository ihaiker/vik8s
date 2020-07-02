package plugins

import (
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"plugin"
)

type (
	ReduceHandler func(item *config.Directive) (obj metav1.Object, handler bool)
	LoadPlugin    func() ReduceHandler
	manager       []ReduceHandler
)

var Manager = manager{}

func (m *manager) Handler(item *config.Directive) (obj metav1.Object, handler bool) {
	for _, reduceHandler := range *m {
		if obj, handler = reduceHandler(item); handler {
			return
		}
	}
	return
}

func plugins() []string {
	files := make([]string, 0)
	pluginDir := tools.Join("plugins", "reduce")
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
		Manager = append(Manager, sl.(LoadPlugin)())
	}
}
