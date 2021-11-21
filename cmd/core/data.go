package core

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

var dataDir = [][]string{
	{"docker", "Docker", "/var/lib/docker"},

	{"kubernetes/config", "Kubernetes config", "/etc/kubernetes"},
	{"kubernetes/data", "Kubernetes data", "/var/lib/kubelet"},
	{"etcd/config", "etcd config", "/etc/etcd"},
	{"etcd/data", "etcd data dir", "/var/lib/etcd"},

	{"openebs", "OpenEBS BasePath for hostPath volumes on Nodes", "/var/openebs"},

	{"glusterfs/config", "GlusterFS Configuration Directory", "/etc/glusterfs"},
	{"glusterfs/data", "GlusterFS working directory", "/var/lib/glusterd"},
	{"glusterfs/heketi", "GlusterFS Heketi Data Folder", "/var/lib/heketi"},

	{"ceph/config", "the ceph cluster config", "/etc/ceph"},
	{"ceph/data", "the ceph cluster data", "/var/lib/ceph"},
}

var mvShell = `
FROM_DIR="%s"
TO_DIR="%s"
if [ -e $FROM_DIR ]; then
  echo "$FROM_DIR dir exists " 
  if [ -L "$FROM_DIR" ]; then
    echo "delete link $FROM_DIR"
    rm -f $FROM_DIR
  elif [ -h "$FROM_DIR" ]; then
    echo "delete hard link $FROM_DIR"
    rm -f $FROM_DIR
  else
    echo "move $FROM_DIR $TO_DIR"
    if [ -e $TO_DIR ]; then
        mv $FROM_DIR/* $TO_DIR
		rmdir $FROM_DIR
    else
        mv $FROM_DIR $TO_DIR
    fi
  fi
fi
mkdir -p $TO_DIR
ln -s $TO_DIR $FROM_DIR
`

var dataCmd = &cobra.Command{
	Use: "data", Short: "Create a data folder link",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return cobra.OnlyValidArgs(cmd, args)
	}, ValidArgs: []string{"all", "*"},
	Long: `
  由于kubernetes系统组件众多，并且很多组件部署位置不一样，通常情况下为了方便管理，会把数据文件放到一起。
  但是如果您这样操作的话在某些情况下出现问题之后的排除上就存在了一些问题。例如教程中的实例的数据位置不一致，在学习或者使用中出现一些问题难以定位。
  为了解决此问题，本程序提供了全局数据位置连接命令（使用硬连接），可以让所有数据存储到一起便于管理。

  Due to the large number of kubernetes system components, and the deployment location of many components are different,
  usually in order to facilitate management, data files will be put together. However, 
  if you do this, there will be some problems in the elimination of the problem in some cases. 
  For example, the data locations of the examples in the tutorial are inconsistent,
  and some problems are difficult to locate during learning or use. In order to solve this problem, 
  this program provides a global data location connection command (using hard connection), 
  which can store all the data together for easy management.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(hosts.Nodes()) == 0 {
			fmt.Println("No nodes found. please use：`vik8s host` add node")
			return
		}
		data, _ := cmd.Flags().GetString("dir")
		dirs := getDir(args)
		for _, node := range hosts.Nodes() {
			for _, dir := range dirs {
				fmt.Printf("%s %s -> %s\n", node.Hostname, dir[2], filepath.Join(data, dir[0]))
				err := node.Shell(fmt.Sprintf(mvShell, dir[2], filepath.Join(data, dir[0])), func(stdout io.Reader) error {
					_, err := io.Copy(os.Stdout, stdout)
					return err
				})
				utils.Panic(err, "error: %s,%s", node.Host, dir[2])
			}
		}
	},
}

func getDir(args []string) [][]string {
	if args[0] == "all" || args[0] == "*" {
		return dataDir
	}

	dirs := make([][]string, 0)
	for _, dir := range dataDir {
		if utils.Search(args, dir[0]) != -1 {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func table() string {
	out := bytes.NewBufferString("")
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"name", "description", "dir"})
	table.SetBorder(true)
	table.SetCenterSeparator("*")
	//table.SetColumnSeparator("╪")
	table.SetRowSeparator("-")
	table.AppendBulk(dataDir)
	table.Render()
	return out.String()
}
func init() {
	dataCmd.Flags().String("dir", "/data", "The default storage location of all data")

	for _, value := range dataDir {
		dataCmd.ValidArgs = append(dataCmd.ValidArgs, value[0])
	}
	dataCmd.Long += table()
	dataCmd.Long += `For example:
    vik8s data --dir=/userdata kubernetes/config
    /etc/kubernetes --> /userdata/kubernetes/config
`
}
