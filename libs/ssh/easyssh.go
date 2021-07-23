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
	"time"
)

type (
	// sshConfig for ssh proxy config
	sshConfig struct {
		User         string
		Server       string
		Key          []byte //pem key bytes
		KeyPath      string //private key path
		Port         string
		Passphrase   string
		Password     string
		Timeout      time.Duration
		Ciphers      []string
		KeyExchanges []string
		Fingerprint  string

		// Enable the use of insecure ciphers and key exchange methods.
		// This enables the use of the the following insecure ciphers and key exchange methods:
		// - aes128-cbc
		// - aes192-cbc
		// - aes256-cbc
		// - 3des-cbc
		// - diffie-hellman-group-exchange-sha256
		// - diffie-hellman-group-exchange-sha1
		// Those algorithms are insecure and may allow plaintext data to be recovered by an attacker.
		UseInsecureCipher bool
	}

	// easySSHConfig Contains main authority information.
	// User field should be a name of user on remote server (ex. john in ssh john@example.com).
	// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
	// Key is a path to private key on your local machine.
	// Port is SSH server port on remote machine.
	// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
	// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
	easySSHConfig struct {
		sshConfig
		Proxy *sshConfig
	}
	StreamWatcher func(stdout io.Reader)
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
func getSSHConfig(config *sshConfig) (*ssh.ClientConfig, io.Closer, error) {
	var sshAgent io.Closer

	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if config.Password != "" {
		auths = append(auths, ssh.Password(config.Password))
	}

	if config.Key != nil {
		if signer, err := getKeySigner(config.Key, config.Passphrase); err != nil {
			return nil, nil, utils.Wrap(err, "get key singer")
		} else {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	} else if config.KeyPath != "" {
		if signer, err := getKeySignerFile(config.KeyPath, config.Passphrase); err != nil {
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
func (conf *easySSHConfig) connect() (*ssh.Session, *ssh.Client, error) {
	var client *ssh.Client
	var err error

	targetConfig, closer, err := getSSHConfig(&conf.sshConfig)
	if err != nil {
		return nil, nil, err
	}
	if closer != nil {
		defer closer.Close()
	}

	// Enable proxy command
	if conf.Proxy != nil {
		proxyConfig, closer, err := getSSHConfig(conf.Proxy)
		if err != nil {
			return nil, nil, err
		}
		if closer != nil {
			defer closer.Close()
		}

		proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(conf.Proxy.Server, conf.Proxy.Port), proxyConfig)
		if err != nil {
			return nil, nil, err
		}

		conn, err := proxyClient.Dial("tcp", net.JoinHostPort(conf.Server, conf.Port))
		if err != nil {
			return nil, nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(conf.Server, conf.Port), targetConfig)
		if err != nil {
			return nil, nil, err
		}

		client = ssh.NewClient(ncc, chans, reqs)
	} else {
		client, err = ssh.Dial("tcp", net.JoinHostPort(conf.Server, conf.Port), targetConfig)
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

func (conf *easySSHConfig) Stream(command string, watch StreamWatcher) (err error) {
	var session *ssh.Session
	var client *ssh.Client

	// connect to remote host
	if session, client, err = conf.connect(); err != nil {
		return
	}
	defer client.Close()
	defer session.Close()

	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	session.Stderr = writer
	session.Stdout = writer
	go watch(reader)
	if err := session.Start(command); err != nil {
		return err
	}
	err = session.Wait()
	return
}

// Run command on remote machine and returns its stdout as a string
func (conf *easySSHConfig) Run(command string) (out []byte, err error) {
	stream := bytes.NewBufferString("")
	if err = conf.Stream(command, func(stdout io.Reader) {
		_, _ = io.Copy(stream, stdout)
	}); err != nil {
		return
	}
	out = stream.Bytes()
	length := len(out)
	if length > 0 && out[length-1] == '\n' {
		out = out[0 : length-1]
	}
	return
}

func (conf *easySSHConfig) Mkdir(path string) error {
	_, err := conf.Run("mkdir -p " + strconv.Quote(path))
	return err
}

func (conf *easySSHConfig) ScpContent(content []byte, destFilePath string) error {
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
func (conf *easySSHConfig) Scp(sourceFilePath string, destFilePath string, bars ...func(step, total int64)) error {
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

// Pull uploads sourceFile to remote machine like native scp console app.
func (conf *easySSHConfig) Pull(sourceFilePath string, destFilePath string, bars ...func(step, total int64)) error {
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
