package certs

import (
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"k8s.io/apimachinery/pkg/util/validation"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	// PrivateKeyBlockType is a possible value for pem.Block.Type.
	PrivateKeyBlockType = "PRIVATE KEY"
	// PublicKeyBlockType is a possible value for pem.Block.Type.
	PublicKeyBlockType = "PUBLIC KEY"
	// CertificateBlockType is a possible value for pem.Block.Type.
	CertificateBlockType = "CERTIFICATE"
	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
	rsaKeySize             = 2048
)

type Config struct {
	*certutil.Config
	CertificateValidity time.Duration
}

func NewConfig(commonName string) *Config {
	return &Config{
		Config: &certutil.Config{
			CommonName: commonName,
		},
		CertificateValidity: time.Now().AddDate(100, 0, 0).Sub(time.Now()),
	}
}

func NewPrivateKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
	utils.Panic(err, "creates an RSA private key")
	return key
}

func WriteKey(pkiPath, name string, key *rsa.PrivateKey) {
	utils.Assert(key != nil, "private key cannot be nil")

	privateKeyPath := pathForKey(pkiPath, name)
	err := keyutil.WriteKey(privateKeyPath, EncodePrivateKeyPEM(key))
	utils.Panic(err, "unable to write private key to file %s", privateKeyPath)
}

// tries to load the key from the disk and validates that it is valid
func TryLoadPrivateKeyFromDisk(pkiPath, name string) *rsa.PrivateKey {
	privateKeyPath := pathForKey(pkiPath, name)

	// Parse the private key from a file
	privKey, err := keyutil.PrivateKeyFromFile(privateKeyPath)
	utils.Panic(err, "couldn't load the private key file %s ", privateKeyPath)

	key, match := privKey.(*rsa.PrivateKey)
	utils.Assert(match, "the private key file %s isn't in RSA format", privateKeyPath)
	return key
}

func TryLoadPublicKeyFromDisk(pkiPath, name string) *rsa.PublicKey {
	publicKeyPath := pathForPublicKey(pkiPath, name)
	// Parse the public key from a file
	pubKeys, err := keyutil.PublicKeysFromFile(publicKeyPath)
	utils.Panic(err, "couldn't load the public key file %s", publicKeyPath)
	p := pubKeys[0].(*rsa.PublicKey)
	return p
}

func NewCertificate(key *rsa.PrivateKey, config *Config) *x509.Certificate {
	now := time.Now()

	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: config.Organization,
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(config.CertificateValidity).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &tmpl, &tmpl, key.Public(), key)
	utils.Panic(err, "unable to create self-signed certificate [%s]", config.CommonName)

	cert, err := x509.ParseCertificate(certDERBytes)
	utils.Panic(err, "unable to create self-signed certificate [%s]", config.CommonName)
	return cert
}

//creates new certificate and private key for the certificate authority
func NewCertificateAuthority(config *Config) (cert *x509.Certificate, key *rsa.PrivateKey) {
	key = NewPrivateKey()
	cert = NewCertificate(key, config)
	return
}

// creates new certificate and key by passing the certificate authority certificate and key
func NewCertAndKey(caCert *x509.Certificate, caKey *rsa.PrivateKey, config *Config) (*x509.Certificate, *rsa.PrivateKey) {
	key := NewPrivateKey()
	cert := NewSignedCert(config, key, caCert, caKey)
	return cert, key
}

// HasServerAuth returns true if the given certificate is a ServerAuth
func HasServerAuth(cert *x509.Certificate) bool {
	for i := range cert.ExtKeyUsage {
		if cert.ExtKeyUsage[i] == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}

func WriteCertAndKey(pkiPath string, name string, cert *x509.Certificate, key *rsa.PrivateKey) {
	WriteKey(pkiPath, name, key)
	WriteCert(pkiPath, name, cert)
}

// WriteCert stores the given certificate at the given location
func WriteCert(pkiPath, name string, cert *x509.Certificate) {
	utils.Assert(cert != nil, "certificate cannot be nil when writing to file")

	certificatePath := pathForCert(pkiPath, name)
	err := certutil.WriteCert(certificatePath, EncodeCertPEM(cert))
	utils.Panic(err, "unable to write certificate to file %s", certificatePath)
}

func WritePublicKey(pkiPath, name string, key crypto.PublicKey) {
	utils.Assert(key != nil, "public key cannot be nil when writing to file")
	publicKeyBytes := EncodePublicKeyPEM(key)
	publicKeyPath := pathForPublicKey(pkiPath, name)
	err := keyutil.WriteKey(publicKeyPath, publicKeyBytes)
	utils.Panic(err, "unable to write public key to file %q", publicKeyPath)
}

func CertOrKeyExist(pkiPath, name string) bool {
	certificatePath, privateKeyPath := PathsForCertAndKey(pkiPath, name)

	_, certErr := os.Stat(certificatePath)
	_, keyErr := os.Stat(privateKeyPath)
	if os.IsNotExist(certErr) && os.IsNotExist(keyErr) {
		// The cert or the key did not exist
		return false
	}

	// Both files exist or one of them
	return true
}

func TryLoadCertAndKeyFromDisk(pkiPath, name string) (*x509.Certificate, *rsa.PrivateKey) {
	cert := TryLoadCertFromDisk(pkiPath, name)
	key := TryLoadPrivateKeyFromDisk(pkiPath, name)
	return cert, key
}

// TryLoadCertFromDisk tries to load the cert from the disk and validates that it is valid
func TryLoadCertFromDisk(pkiPath, name string) *x509.Certificate {
	certificatePath := pathForCert(pkiPath, name)
	certs, err := certutil.CertsFromFile(certificatePath)
	utils.Panic(err, "couldn't load the certificate file %s", certificatePath)

	cert := certs[0]

	// Check so that the certificate is valid now
	now := time.Now()
	utils.Assert(!now.Before(cert.NotBefore), "the certificate is not valid yet")
	utils.Assert(!now.After(cert.NotAfter), "the certificate has expired")
	return cert
}

func TryLoadPrivatePublicKeyFromDisk(pkiPath, name string) (*rsa.PrivateKey, *rsa.PublicKey) {
	priKey := TryLoadPrivateKeyFromDisk(pkiPath, name)
	pubKey := TryLoadPublicKeyFromDisk(pkiPath, name)
	return priKey, pubKey
}

func PathsForCertAndKey(pkiPath, name string) (string, string) {
	return pathForCert(pkiPath, name), pathForKey(pkiPath, name)
}

func pathForCert(pkiPath, name string) string {
	return filepath.Join(pkiPath, fmt.Sprintf("%s.crt", name))
}

func pathForKey(pkiPath, name string) string {
	return filepath.Join(pkiPath, fmt.Sprintf("%s.key", name))
}

func pathForPublicKey(pkiPath, name string) string {
	return filepath.Join(pkiPath, fmt.Sprintf("%s.pub", name))
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) *x509.Certificate {
	serial, _ := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))

	utils.Assert(cfg.CommonName != "", "must specify a CommonName")
	utils.Assert(len(cfg.Usages) > 0, "must specify at least one ExtKeyUsage")

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(cfg.CertificateValidity).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	utils.Panic(err, "unable to sign certificate ")

	cert, err := x509.ParseCertificate(certDERBytes)
	utils.Panic(err, "unable to sign certificate ")
	return cert
}

func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  CertificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}
func EncodePublicKeyPEM(key crypto.PublicKey) []byte {
	der, err := x509.MarshalPKIXPublicKey(key)
	utils.Panic(err, "encode public key to pem")

	block := pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: der,
	}
	return pem.EncodeToMemory(&block)
}
func EncodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  RSAPrivateKeyBlockType, // "RSA PRIVATE KEY"
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

func GetAltNames(SANs []string, certName string) *certutil.AltNames {
	// create AltNames with defaults DNSNames/IPs
	altNames := &certutil.AltNames{}

	for _, altname := range SANs {
		if ip := net.ParseIP(altname); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else if len(validation.IsDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else {
			fmt.Printf(
				"[certificates] WARNING: '%s' was not added to the '%s' SAN, because it is not a valid IP or RFC-1123 compliant DNS entry\n",
				altname,
				certName,
			)
		}
	}
	return altNames
}
