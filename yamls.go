package main

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

//-- go:generate go get -u github.com/shuLhan/go-bindata/cmd/go-bindata
//go:generate go run yamls.go ${FILE}
//go:generate go-bindata -modtime 1590460659 -pkg yamls -o yaml/assets.go -ignore .*\.go -ignore .*\.part yaml/...
//go:generate go fmt yaml/assets.go

//碎片文件合并
func main() {
	dir, err := filepath.Abs(os.Args[1])
	utils.Panic(err, "path not found")
	_ = filepath.Walk(filepath.Join(dir, "yaml"), func(path string, info os.FileInfo, err error) error {

		if info.IsDir() && filepath.Ext(path) == ".part" {
			mergeFileName := path[0:len(path)-5] + ".yaml"

			dir, _ := os.Open(path)
			names, _ := dir.Readdirnames(0)
			sort.Strings(names)

			writer := bytes.NewBufferString("")
			for _, name := range names {
				if filepath.Ext(name) != ".yaml" {
					continue
				}
				writer.WriteString(fmt.Sprintf("# name %s\n---\n", name))
				bs, _ := ioutil.ReadFile(filepath.Join(path, name))
				writer.Write(bs)
				writer.WriteString("\n\n")
			}
			return ioutil.WriteFile(mergeFileName, writer.Bytes(), 0644)
		}
		return nil
	})
}
