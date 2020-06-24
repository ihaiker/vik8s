package etcdcerts

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/libs/utils"
	"time"
)

func line(name, format string, params ...interface{}) {
	fmt.Printf("[cert][etcd][%s] %s \n", name, fmt.Sprintf(format, params...))
}

type createAction func(name, dir string, sans []string, certificateValidity time.Duration)

func CreatePKIAssets(name, dir string, sans []string, certificateValidity time.Duration) {
	line(name, "creating PKI assets %s", dir)

	actions := []createAction{
		createCACertAndKeyFiles,
		createServerCertAndKeyFiles,
		createPeerCertAndKeyFiles,
		createCtlClientCertAndKeyFiles,
		createAPIServerEtcdClientCertAndKeyFiles,
		createKubeHealthcheckClientFiles,
	}
	for _, action := range actions {
		action(name, dir, sans, certificateValidity)
	}
	line(name, "valid certificates and keys now exist in %q", dir)
}

func createCACertAndKeyFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	if certs.CertOrKeyExist(dir, "ca") {
		return
	}
	line(name, "creating a self signed etcd CA certificate and key files")
	cfg := certs.NewConfig(name)
	cfg.CertificateValidity = certificateValidity
	etcdCACert, etcdCAKey := certs.NewCertificateAuthority(cfg)
	certs.WriteCertAndKey(dir, "ca", etcdCACert, etcdCAKey)
}

func createServerCertAndKeyFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	line(name, "creating a new certificate and key files for etcd server")
	caCert, caKey := loadCaCertAndKey(dir)

	altNames := certs.GetAltNames(sans, "server")
	config := certs.NewConfig(name)
	{
		config.AltNames = *altNames
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}

	serverCert, serverKey := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, fmt.Sprintf("server-%s", name), serverCert, serverKey)
}

func createPeerCertAndKeyFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	line(name, "creating a new certificate and key files for etcd peering")
	caCert, caKey := loadCaCertAndKey(dir)
	altNames := certs.GetAltNames(sans, "peer")
	config := certs.NewConfig(name)
	{
		config.AltNames = *altNames
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}
	peerCert, peerKey := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, fmt.Sprintf("peer-%s", name), peerCert, peerKey)
}

func createCtlClientCertAndKeyFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	if certs.CertOrKeyExist(dir, "etcdctl-etcd-client") {
		return
	}

	line(name, "creating a new client certificate for the etcdctl")
	caCert, caKey := loadCaCertAndKey(dir)

	commonName := fmt.Sprintf("%s-etcdctl", name)
	// MastersGroup defines the well-known group for the apiservers. This group is also superuser by default
	// (i.e. bound to the cluster-admin ClusterRole)
	organization := "system:masters"
	config := certs.NewConfig(commonName)
	{
		config.Organization = []string{organization}
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}
	cert, key := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, "etcdctl-etcd-client", cert, key)
}

func createAPIServerEtcdClientCertAndKeyFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	if certs.CertOrKeyExist(dir, "apiserver-etcd-client") {
		return
	}

	line(name, "creating a new client certificate for the apiserver calling etcd")
	caCert, caKey := loadCaCertAndKey(dir)
	// MastersGroup defines the well-known group for the apiservers. This group is also superuser by default
	// (i.e. bound to the cluster-admin ClusterRole)
	organization := "system:masters"
	config := certs.NewConfig("kube-apiserver-etcd-client")
	{
		config.Organization = []string{organization}
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}
	cert, key := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, "apiserver-etcd-client", cert, key)
}

func createKubeHealthcheckClientFiles(name, dir string, sans []string, certificateValidity time.Duration) {
	if certs.CertOrKeyExist(dir, "healthcheck-client") {
		return
	}
	caCert, caKey := loadCaCertAndKey(dir)
	line(name, "creating healthcheck-client")

	config := certs.NewConfig("kube-etcd-healthcheck-client")
	{
		config.Organization = []string{"system:masters"}
		config.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		config.CertificateValidity = certificateValidity
	}
	cert, key := certs.NewCertAndKey(caCert, caKey, config)
	certs.WriteCertAndKey(dir, "healthcheck-client", cert, key)
}

// Checks if certificate authority exists in the PKI directory
func loadCaCertAndKey(dir string) (*x509.Certificate, *rsa.PrivateKey) {
	utils.Assert(certs.CertOrKeyExist(dir, "ca"), "couldn't load ca certificate authority from %s", dir)
	return certs.TryLoadCertAndKeyFromDisk(dir, "ca")
}
