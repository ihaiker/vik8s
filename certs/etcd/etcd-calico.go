package etcdcerts

import (
	"crypto/x509"
	"github.com/ihaiker/vik8s/certs"
	"time"
)

func CreateCalicoETCDPKIAssert(dir string, certificateValidity time.Duration) (certPath, keyPath string) {
	certPath, keyPath = certs.PathsForCertAndKey(dir, "calico-etcd-client")

	if certs.CertOrKeyExist(dir, "calico-etcd-client") {
		return
	}

	line("caolico-etcd-client", "create caolico-etcd-client.crt")
	caCert, caKey := loadCaCertAndKey(dir)
	config := certs.NewConfig("calico-etcd-client")
	{
		config.Organization = []string{"system:masters"}
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}

	cert, key := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, "calico-etcd-client", cert, key)
	return
}
