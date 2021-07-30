package dockercert

import (
	"crypto/x509"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
)

func GenerateBootstrapCertificates(dir string) (cfg *config.DockerCertsConfiguration, err error) {
	cfg = &config.DockerCertsConfiguration{Custom: false, Enable: true}

	if cfg.CaCertPath, cfg.CaPrivateKeyPath = certs.PathsForCertAndKey(dir, "ca"); utils.NotExists(cfg.CaCertPath) || utils.NotExists(cfg.CaPrivateKeyPath) {
		config := certs.NewConfig("ca")
		cert, key := certs.NewCertificateAuthority(config)
		certs.WriteCertAndKey(dir, "ca", cert, key)
	}
	caCert, caKey := certs.TryLoadCertAndKeyFromDisk(dir, "ca")

	if cfg.ClientCertPath, cfg.ClientKeyPath = certs.PathsForCertAndKey(dir, "client"); utils.NotExists(cfg.ClientCertPath) || utils.NotExists(cfg.ClientKeyPath) {
		config := certs.NewConfig("client")
		{
			config.Organization = []string{"system:client"}
			config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
			config.AltNames = *certs.GetAltNames([]string{}, "client")
		}
		clientCert, clientKey := certs.NewCertAndKey(caCert, caKey, config)
		certs.WriteCertAndKey(dir, "client", clientCert, clientKey)
	}
	return
}

func GenerateServerCertificates(node *ssh.Node, options *config.DockerCertsConfiguration) (serverCertPath, serverKeyPath string, err error) {
	dir := filepath.Dir(options.CaCertPath)
	caCert, caKey := certs.TryLoadCertAndKeyFromDisk(dir, "ca")

	if serverCertPath, serverKeyPath = certs.PathsForCertAndKey(dir, "server-"+node.Host); utils.NotExists(serverCertPath) || utils.NotExists(serverKeyPath) {
		config := certs.NewConfig("server")
		{
			config.Organization = []string{"system:server"}
			config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
			config.AltNames = *certs.GetAltNames([]string{node.Host, "127.0.0.1", "localhost"}, "server")
		}
		serverCert, serverKey := certs.NewCertAndKey(caCert, caKey, config)
		certs.WriteCertAndKey(dir, "server-"+node.Host, serverCert, serverKey)
	}
	return
}
