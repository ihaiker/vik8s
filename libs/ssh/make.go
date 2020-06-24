package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"sync"
)

func Sync(nodes []*Node, run func(i int, node *Node)) {
	hasError := ""
	gw := new(sync.WaitGroup)
	for i, node := range nodes {
		gw.Add(1)
		go func(i int, node *Node) {
			defer gw.Done()
			defer utils.Catch(func(err error) {
				hasError += fmt.Sprintf("%s %s\n", node.Host, err.Error())
			})
			run(i, node)
		}(i, node)
	}
	gw.Wait()
	utils.Assert(hasError == "", hasError)
}
