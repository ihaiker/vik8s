package ssh

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (node *Node) SudoScp(localPath, remotePath string) error {
	path := fmt.Sprintf("/tmp/vik8s/%s", filepath.Base(localPath))

	if err := node.scp(localPath, path, true); err != nil {
		return err
	}
	dir := filepath.Dir(remotePath)
	if err := node.SudoCmd("mkdir -p " + dir); err != nil {
		return err
	}
	return node.SudoCmd(fmt.Sprintf("mv %s %s", strconv.Quote(path), strconv.Quote(remotePath)))
}

func (node *Node) SudoScpContent(content []byte, remotePath string) error {
	path := fmt.Sprintf("/tmp/vik8s/%s", filepath.Base(remotePath))

	node.Logger("scp content to %s", path)
	if err := node.ScpContent(content, path); err != nil {
		return err
	}
	dir := filepath.Dir(remotePath)
	if err := node.SudoCmd("mkdir -p " + dir); err != nil {
		return err
	}
	return node.SudoCmd(fmt.Sprintf("mv %s %s", strconv.Quote(path), strconv.Quote(remotePath)))
}

func (node *Node) Scp(localPath, remotePath string) error {
	return node.scp(localPath, remotePath, true)
}

func (node *Node) ScpContent(content []byte, remotePath string) error {
	return node.easyssh().ScpContent(content, remotePath)
}

func (node *Node) scp(localPath, remotePath string, showProgressBar bool) error {

	var bar *pb.ProgressBar
	if showProgressBar {
		bar = pb.New64(100)
		defer bar.Finish()
		bar.SetWriter(os.Stdout)
		bar.SetRefreshRate(time.Millisecond * 300)
		bar.Set("prefix",
			fmt.Sprintf("%s scp %s %s  ", node.Prefix(), localPath, remotePath))
		bar.Start()
	}

	return node.easyssh().Scp(localPath, remotePath, func(step, all int64) {
		if showProgressBar {
			bar.SetTotal(all)
			bar.SetCurrent(step)
		}
	})
}

func (node *Node) Pull(remotePath, localPath string) error {
	node.Logger("pull %s %s", remotePath, localPath)

	bar := pb.New64(100)
	defer bar.Finish()
	bar.SetWriter(os.Stdout)
	bar.SetRefreshRate(time.Millisecond * 300)
	bar.Set(pb.Terminal, true)
	bar.Start()

	return node.easyssh().Pull(remotePath, localPath, func(step, total int64) {
		bar.SetTotal(total)
		bar.SetCurrent(step)
	})
}
