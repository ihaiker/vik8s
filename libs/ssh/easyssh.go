package ssh

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type (
	StreamWatcher func(stdout io.Reader) error
)

// returns ssh.Signer from user you running app home path + cutted key path.
func getKeySignerFile(keypath, passphrase string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}
	return getKeySigner(buf, passphrase)
}

func getKeySigner(key []byte, passphrase string) (pubkey ssh.Signer, err error) {
	if passphrase != "" {
		pubkey, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	} else {
		pubkey, err = ssh.ParsePrivateKey(key)
	}
	return pubkey, err
}

// returns *ssh.ClientConfig and io.Closer.
// if io.Closer is not nil, io.Closer.Close() should be called when
// *ssh.ClientConfig is no longer used.
func getSSHConfig(config *Node) (*ssh.ClientConfig, io.Closer, error) {
	var sshAgent io.Closer

	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if config.Password != "" {
		auths = append(auths, ssh.Password(config.Password))
	}

	if config.PrivateKeyRaw != "" {
		if signer, err := getKeySigner([]byte(config.PrivateKey), config.Passphrase); err != nil {
			return nil, nil, utils.Wrap(err, "get key singer")
		} else {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	} else if config.PrivateKey != "" {
		if signer, err := getKeySignerFile(config.PrivateKey, config.Passphrase); err != nil {
			return nil, nil, utils.Wrap(err, "get key singer error: %v", err)
		} else {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	c := ssh.Config{}
	if config.UseInsecureCipher {
		c.SetDefaults()
		c.Ciphers = append(c.Ciphers, "aes128-cbc", "aes192-cbc", "aes256-cbc", "3des-cbc")
		c.KeyExchanges = append(c.KeyExchanges, "diffie-hellman-group-exchange-sha1", "diffie-hellman-group-exchange-sha256")
	}

	if len(config.Ciphers) > 0 {
		c.Ciphers = append(c.Ciphers, config.Ciphers...)
	}

	if len(config.KeyExchanges) > 0 {
		c.KeyExchanges = append(c.KeyExchanges, config.KeyExchanges...)
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if config.Fingerprint != "" {
		hostKeyCallback = func(hostname string, remote net.Addr, publicKey ssh.PublicKey) error {
			if ssh.FingerprintSHA256(publicKey) != config.Fingerprint {
				return fmt.Errorf("ssh: host key fingerprint mismatch")
			}
			return nil
		}
	}
	clientConfig := &ssh.ClientConfig{
		Config:          c,
		Timeout:         config.Timeout,
		User:            config.User,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
	}
	return clientConfig, sshAgent, nil
}

// connect to remote server using easySSHConfig struct and returns *ssh.Session
func (node *Node) connect() (*ssh.Session, *ssh.Client, error) {
	var client *ssh.Client
	var err error

	targetConfig, closer, err := getSSHConfig(node)
	if err != nil {
		return nil, nil, err
	}
	if closer != nil {
		defer closer.Close()
	}

	// Enable proxy command
	if node.Proxy != "" {
		proxyConfig, closer, err := getSSHConfig(node.ProxyNode)
		if err != nil {
			return nil, nil, err
		}
		if closer != nil {
			defer closer.Close()
		}

		proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(node.ProxyNode.Host, strconv.Itoa(node.ProxyNode.Port)), proxyConfig)
		if err != nil {
			return nil, nil, err
		}

		conn, err := proxyClient.Dial("tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
		if err != nil {
			return nil, nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(node.Host, strconv.Itoa(node.Port)), targetConfig)
		if err != nil {
			return nil, nil, err
		}

		client = ssh.NewClient(ncc, chans, reqs)
	} else {
		client, err = ssh.Dial("tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)), targetConfig)
		if err != nil {
			return nil, nil, err
		}
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}
	return session, client, nil
}

func (conf *Node) stream(command string, watch StreamWatcher) (err error) {
	var session *ssh.Session
	var client *ssh.Client

	// connect to remote host
	if session, client, err = conf.connect(); err != nil {
		return
	}
	defer client.Close()
	defer session.Close()

	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	watchErr := make(chan error)
	defer close(watchErr)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				watchErr <- e.(error)
			}
		}()
		watchErr <- watch(stdout)
	}()

	if err = session.Start(command); err != nil {
		return
	}
	if err = session.Wait(); err != nil {
		errBytes, _ := ioutil.ReadAll(stderr)
		if exitErr, match := err.(*ssh.ExitError); match {
			if len(errBytes) > 0 {
				err = utils.Wrap(err, string(utils.Trdn(errBytes)))
			} else {
				err = utils.Error("exit %v: %v", exitErr.ExitStatus(), string(utils.Trdn([]byte(exitErr.Msg()))))
			}
		}
	}
	if err, has := <-watchErr; has && err != nil {
		return err
	}
	return
}

// Run command on remote machine and returns its stdout as a string
func (conf *Node) run(command string) (out []byte, err error) {
	stream := bytes.NewBufferString("")
	if err = conf.stream(command, func(stdout io.Reader) error {
		_, e := io.Copy(stream, stdout)
		return e
	}); err == nil {
		out = utils.Trdn(stream.Bytes())
	}
	return
}

func (conf *Node) mkdir(path string) error {
	_, err := conf.run("mkdir -p " + strconv.Quote(path))
	return err
}

func (conf *Node) scpContent(content []byte, destFilePath string) error {
	session, client, err := conf.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	if sftpClient, err := sftp.NewClient(client); err != nil {
		return utils.Wrap(err, "open sftp client")
	} else {
		dir := filepath.Dir(destFilePath)
		if err := sftpClient.MkdirAll(dir); err != nil {
			return utils.Wrap(err, "mkdir remote folder %s", dir)
		}

		dstFile, err := sftpClient.Create(destFilePath)
		if err != nil {
			return utils.Wrap(err, "create remote file %s", destFilePath)
		}
		defer dstFile.Close()

		_, err = dstFile.Write(content)
	}
	return err
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (conf *Node) _scp(sourceFilePath string, destFilePath string, bars ...func(step, total int64)) error {
	if utils.NotExists(sourceFilePath) {
		return utils.Error("file not found %s", sourceFilePath)
	}
	session, client, err := conf.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	if sftpClient, err := sftp.NewClient(client); err != nil {
		return utils.Wrap(err, "open sftp client")
	} else {
		srcFile, _ := os.Open(sourceFilePath)
		defer srcFile.Close()

		dir := filepath.Dir(destFilePath)
		if err := sftpClient.MkdirAll(dir); err != nil {
			return utils.Wrap(err, "mkdir remote folder %s", dir)
		}

		dstFile, err := sftpClient.Create(destFilePath)
		if err != nil {
			return utils.Wrap(err, "create remote file %s", destFilePath)
		}
		defer dstFile.Close()

		//processing
		bs := make([]byte, 1024)
		var length int
		fs, _ := srcFile.Stat()
		total := fs.Size()
		var step int64

		for {
			if length, err = srcFile.Read(bs); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if _, err = dstFile.Write(bs[0:length]); err != nil {
				return err
			}
			step += int64(length)
			for _, bar := range bars {
				bar(step, total)
			}
		}
	}
	return err
}

// Pull uploads sourceFile to remote machine like native scp console app.
func (conf *Node) pull(sourceFilePath string, destFilePath string, bars ...func(step, total int64)) error {
	//mkdir local dir
	if utils.NotExists(filepath.Dir(destFilePath)) {
		if err := utils.Mkdir(filepath.Dir(destFilePath)); err != nil {
			return err
		}
	}

	session, client, err := conf.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	if sftpClient, err := sftp.NewClient(client); err != nil {
		return utils.Wrap(err, "open sftp client")
	} else {

		srcFile, err := sftpClient.Open(sourceFilePath)
		if err != nil {
			return utils.Wrap(err, "remote file %s not found", sourceFilePath)
		}
		defer srcFile.Close()

		destFile, err := os.Open(destFilePath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		//processing
		var bs []byte
		var length int
		fs, _ := srcFile.Stat()
		total := fs.Size()
		var step int64

		for {
			if length, err = srcFile.Read(bs); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			step += int64(length)
			for _, bar := range bars {
				bar(step, total)
			}
		}
	}
	return err
}
