package kubecerts

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
	"time"
)

//内部
//for etcd
//etcd/healthcheck-client
//apiserver-etcd-client

type Node struct {
	Name string
	Host string

	ApiServer string
	SvcCIDR   string

	CertificateValidity time.Duration

	SANS []string
}

func line(name, format string, params ...interface{}) {
	logs.Infof("[cert][k8s][%s] %s", name, fmt.Sprintf(format, params...))
}

type createAction func(dir string, node Node)

func CreatePKIAssets(dir string, node Node) {
	line(node.Name, "creating PKI assets %s", dir)
	actions := []createAction{
		createServiceAccountKeyPair,
		createCACertAndKeyFiles,
		createFrontProxyFiles,

		createApiServerFiles,
		createApiServerKubeletClientFiles,
		createFrontProxyClient,
	}

	for _, action := range actions {
		action(dir, node)
	}
	line(node.Name, "valid certificates and keys now exist in %q", dir)
}

func createServiceAccountKeyPair(dir string, node Node) {
	if utils.Exists(filepath.Join(dir, "sa.key")) {
		line(node.Name, "sa.key sa.pub already exist")
		return
	}
	key := certs.NewPrivateKey()
	pub := key.Public()
	certs.WriteKey(dir, "sa", key)
	certs.WritePublicKey(dir, "sa", pub)
}

func createCACertAndKeyFiles(dir string, node Node) {
	if certs.CertOrKeyExist(dir, "ca") {
		line(node.Name, "ca.key and ca.crt already exist")
		return
	}
	line(node.Name, "creating a self signed CA certificate and key files")
	cfg := certs.NewConfig(ClusterName)
	cfg.CertificateValidity = node.CertificateValidity
	caCert, caKey := certs.NewCertificateAuthority(cfg)
	certs.WriteCertAndKey(dir, "ca", caCert, caKey)
}

func createFrontProxyFiles(dir string, node Node) {
	if certs.CertOrKeyExist(dir, "front-proxy-ca") {
		line(node.Name, "front-proxy-ca.key and front-proxy-ca.crt already exist")
		return
	}
	line(node.Name, "creating front-proxy-ca")

	cfg := certs.NewConfig("front-proxy-ca")
	cfg.CertificateValidity = node.CertificateValidity
	caCert, caKey := certs.NewCertificateAuthority(cfg)
	certs.WriteCertAndKey(dir, "front-proxy-ca", caCert, caKey)
}

func createApiServerFiles(dir string, node Node) {
	fileName := "apiserver-" + node.Name
	if certs.CertOrKeyExist(dir, fileName) {
		line(node.Name, "%s.key and %s.crt already exist", fileName, fileName)
		return
	}
	line(node.Name, "creating apiserver %s", node.Name)

	apiServerVip := config.K8S().ApiServerVIP
	mix, _ := tools.AddressRange(node.SvcCIDR)

	cfg := certs.NewConfig("kube-apiserver")
	cfg.CertificateValidity = node.CertificateValidity
	sans := []string{
		"cluster.local", "localhost", node.Name, node.ApiServer,
		"kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster.local",
		"127.0.0.1", node.Host, utils.NextIP(mix).String() /*服务cidr第一个地址为内部通讯使用*/, apiServerVip,
	}
	sans = append(sans, node.SANS...)
	cfg.AltNames = *certs.GetAltNames(sans, "apiserver")
	cfg.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	caCert, caKey := loadCaCertAndKey(dir, "ca")
	cert, key := certs.NewCertAndKey(caCert, caKey, cfg)
	certs.WriteCertAndKey(dir, fileName, cert, key)
}

func createApiServerKubeletClientFiles(dir string, node Node) {
	fileName := "apiserver-kubelet-client-" + node.Name
	if certs.CertOrKeyExist(dir, fileName) {
		line(node.Name, "%s.key and %s.crt already exist", fileName, fileName)
		return
	}
	caCert, caKey := loadCaCertAndKey(dir, "ca")

	line(node.Name, "creating %s", fileName)
	cfg := certs.NewConfig("kube-apiserver-kubelet-client")
	{
		cfg.Organization = []string{"system:masters"}
		cfg.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		cfg.CertificateValidity = node.CertificateValidity
	}
	cert, key := certs.NewCertAndKey(caCert, caKey, cfg)
	certs.WriteCertAndKey(dir, fileName, cert, key)
}

func createFrontProxyClient(dir string, node Node) {
	fileName := "front-proxy-client-" + node.Name
	if certs.CertOrKeyExist(dir, fileName) {
		line(node.Name, "%s.key and %s.crt already exist", fileName, fileName)
		return
	}
	caCert, caKey := loadCaCertAndKey(dir, "front-proxy-ca")
	line(node.Name, "creating %s", fileName)
	cfg := certs.NewConfig("front-proxy-client")
	{
		cfg.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		cfg.CertificateValidity = node.CertificateValidity
	}
	cert, key := certs.NewCertAndKey(caCert, caKey, cfg)
	certs.WriteCertAndKey(dir, fileName, cert, key)
}

// Checks if certificate authority exists in the PKI directory
func loadCaCertAndKey(dir, name string) (*x509.Certificate, *rsa.PrivateKey) {
	utils.Assert(certs.CertOrKeyExist(dir, name), "couldn't load %s/%s", dir, name)
	return certs.TryLoadCertAndKeyFromDisk(dir, name)
}
