package ssh

import (
	"github.com/cheggaaa/pb"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"time"
)

func (node *Node) ScpProgress(localPath, remotePath string) error {
	return node.scp(localPath, remotePath, true)
}

func (node *Node) Scp(localPath, remotePath string) error {
	return node.scp(localPath, remotePath, false)
}

func (node *Node) ScpContent(content []byte, remotePath string) error {
	return node.easyssh().ScpContent(content, remotePath)
}

func (node *Node) scp(localPath, remotePath string, showProgressBar bool) error {
	node.Logger("scp %s %s", localPath, remotePath)

	var bar *pb.ProgressBar
	if showProgressBar {
		bar = pb.New64(100)
		defer bar.Finish()
		bar.SetWriter(os.Stdout)
		bar.SetRefreshRate(time.Millisecond * 300)
		bar.Set(pb.Terminal, true)
		bar.Start()
	}

	return node.easyssh().Scp(localPath, remotePath, func(step, all int64) {
		if showProgressBar {
			bar.SetTotal(all)
			bar.SetCurrent(step)
		}
	})
}

func (node *Node) MustScpContent(content []byte, remotePath string) {
	utils.Panic(node.ScpContent(content, remotePath), "scp %s", remotePath)
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
