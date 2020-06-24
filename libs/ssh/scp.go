package ssh

import (
	"bytes"
	"github.com/cheggaaa/pb"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"time"
)

func (node *Node) ScpProgress(localPath, remotePath string) error {
	return node.scp(localPath, remotePath, true)
}

func (node *Node) Scp(localPath, remotePath string) error {
	return node.scp(localPath, remotePath, false)
}

func (node *Node) Pull(remotePath, localPath string) error {
	node.Logger("pull %s %s", remotePath, localPath)
	return node.connect(func(client *ssh.Client) error {
		if sftpClient, err := sftp.NewClient(client); err != nil {
			return utils.Wrap(err, "open sftp client")
		} else {
			dir := filepath.Dir(localPath)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return utils.Wrap(err, "mkdir local folder %s", dir)
			}

			dstFile, err := sftpClient.Open(remotePath)
			if err != nil {
				return utils.Wrap(err, "create remote file %s", remotePath)
			}
			defer dstFile.Close()

			f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return utils.Wrap(err, "create local file %s", localPath)
			}
			defer f.Close()

			_, err = io.Copy(f, dstFile)
			return err
		}
	})
}

func (node *Node) scp(localPath, remotePath string, showbar bool) error {
	node.Logger("scp %s %s", localPath, remotePath)

	return node.connect(func(client *ssh.Client) error {
		if sftpClient, err := sftp.NewClient(client); err != nil {
			return utils.Wrap(err, "open sftp client")
		} else {
			stat, err := os.Stat(localPath)
			if err != nil {
				return utils.Wrap(err, "open local file %s", localPath)
			}
			fileSize := stat.Size()

			srcFile, err := os.Open(localPath)
			if err != nil {
				return utils.Wrap(err, "open local file %s", localPath)
			}
			defer srcFile.Close()

			dir := filepath.Dir(remotePath)
			if err := node.Mkdir(dir); err != nil {
				return utils.Wrap(err, "mkdir remote folder %s", dir)
			}
			dstFile, err := sftpClient.Create(remotePath)
			if err != nil {
				return utils.Wrap(err, "create remote file %s", remotePath)
			}
			defer dstFile.Close()

			if showbar {
				return node.bar(srcFile, dstFile, fileSize)
			} else {
				_, err = io.Copy(dstFile, srcFile)
				return err
			}
		}
	})
}

func (node *Node) ScpContent(content []byte, remotePath string) error {
	return node.connect(func(client *ssh.Client) error {
		if sftpClient, err := sftp.NewClient(client); err != nil {
			return utils.Wrap(err, "open sftp client")
		} else {
			dir := filepath.Dir(remotePath)
			utils.Panic(node.Mkdir(dir), "mkdir remote folder %s", dir)

			srcFile := bytes.NewBuffer(content)
			dstFile, err := sftpClient.Create(remotePath)
			if err != nil {
				return utils.Wrap(err, "create remote file %s", remotePath)
			}
			defer dstFile.Close()
			_, err = io.Copy(dstFile, srcFile)
			return err
		}
	})
}

func (node *Node) MustScpContent(content []byte, remotePath string) {
	utils.Panic(node.ScpContent(content, remotePath), "scp %s", remotePath)
}

func (node *Node) bar(srcFile io.Reader, dstFile io.Writer, fileSize int64) error {
	// create bar
	bar := pb.New64(fileSize)
	defer bar.Finish()

	bar.SetWriter(os.Stdout)
	bar.SetRefreshRate(time.Second)
	bar.Set(pb.Bytes, true)
	bar.Start()
	pr := bar.NewProxyReader(srcFile)
	_, err := io.Copy(dstFile, pr)
	return err
}
