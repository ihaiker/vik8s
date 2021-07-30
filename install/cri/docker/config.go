package docker

import (
	dockercerts "github.com/ihaiker/vik8s/certs/docker"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"path/filepath"
	"strings"
)

const DockerConfigPath = "docker"
const DockerCertsPath = DockerConfigPath + "/certs.d"

func dockerGenCert(this *config.DockerConfiguration) (err error) {
	if this.TLS != nil && this.TLS.Enable {
		destDir := paths.Join(DockerCertsPath)

		//用户使用自己定义的证书文件
		this.TLS.Custom = this.TLS.CaCertPath != ""

		if !this.TLS.Custom {
			utils.Logs("generator docker certs")
			if this.TLS, err = dockercerts.GenerateBootstrapCertificates(destDir); err != nil {
				return err
			}
			this.TLS.Enable = true
		}

		if !strings.HasPrefix(this.TLS.CaCertPath, destDir) {
			utils.Logs("copy certificates files to .vik8s")
			if this.TLS.CaCertPath, err = utils.Copyto(this.TLS.CaCertPath, destDir); err != nil {
				return
			}
			if this.TLS.ClientKeyPath, err = utils.Copyto(this.TLS.ClientKeyPath, destDir); err != nil {
				return
			}
			if this.TLS.ClientCertPath, err = utils.Copyto(this.TLS.ClientCertPath, destDir); err != nil {
				return
			}
		}
		_ = os.Symlink(this.TLS.CaCertPath, filepath.Join(destDir, "ca.pem"))
		_ = os.Symlink(this.TLS.ClientKeyPath, filepath.Join(destDir, "key.pem"))
		_ = os.Symlink(this.TLS.ClientCertPath, filepath.Join(destDir, "cert.pem"))
	}
	return
}

func Config(this *config.DockerConfiguration) error {
	//copy the daemon.json to local storage.
	if this.DaemonJson != "" {
		destDaemonJson := paths.Join(DockerConfigPath, "daemon.json")
		if err := os.MkdirAll(filepath.Dir(destDaemonJson), os.ModePerm); err != nil {
			return err
		}
		if destDaemonJson != this.DaemonJson {
			utils.Panic(utils.Copy(this.DaemonJson, destDaemonJson), "copy file")
			this.DaemonJson = destDaemonJson
		}
		return nil
	}
	return dockerGenCert(this)
}
