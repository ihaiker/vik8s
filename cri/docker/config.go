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

		this.TLS.Custom = this.TLS.CaCert != ""

		if !this.TLS.Custom {
			utils.Logs("generator docker certs")
			if this.TLS, err = dockercerts.GenerateCertificate(destDir); err != nil {
				return err
			}
			this.TLS.Enable = true
		}

		if !strings.HasPrefix(this.TLS.CaCert, destDir) {
			utils.Logs("copys cert files to .vik8s")
			caPath := filepath.Join(destDir, "ca.pem")
			if err := utils.Copy(this.TLS.CaCert, caPath); err != nil {
				return err
			}
			this.TLS.CaCert = caPath

			certPath := filepath.Join(destDir, "cert.pem")
			if err := utils.Copy(this.TLS.ServerCert, certPath); err != nil {
				return err
			}
			this.TLS.ServerCert = certPath

			keyPath := filepath.Join(destDir, "key.pem")
			if err := utils.Copy(this.TLS.ServerKey, keyPath); err != nil {
				return err
			}
			this.TLS.ServerKey = keyPath
		}
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
