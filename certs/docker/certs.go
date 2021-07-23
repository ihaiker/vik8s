package docker

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/ihaiker/vik8s/config"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func GenerateCertificate(dest string) (*config.DockerCerts, error) {
	org := "vik8s"
	bits := 2048
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return nil, err
	}

	certs := &config.DockerCerts{
		CaCert:       filepath.Join(dest, "ca.pem"),
		CaPrivateKey: filepath.Join(dest, "ca-key.pem"),
		ServerCert:   filepath.Join(dest, "server.pem"),
		ServerKey:    filepath.Join(dest, "server-key.pem"),
		ClientCert:   filepath.Join(dest, "cert.pem"),
		ClientKey:    filepath.Join(dest, "key.pem"),
	}

	if err := generateCACertificate(certs.CaCert, certs.CaPrivateKey, org, bits); err != nil {
		return nil, err
	}

	if err := generateCert(org, certs.CaCert, certs.CaPrivateKey,
		certs.ServerCert, certs.ServerKey, bits, []string{}); err != nil {
		return nil, err
	}

	if err := generateCert(org, certs.CaCert, certs.CaPrivateKey,
		certs.ClientCert, certs.ClientKey, bits, []string{}); err != nil {
		return nil, err
	}
	return certs, nil
}

// generateCACertificate generates a new certificate authority from the specified org
// and bit size and stores the resulting certificate and key file
// in the arguments.
func generateCACertificate(certFile, keyFile, org string, bits int) error {
	template, err := newCertificate(org)
	if err != nil {
		return err
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign
	template.KeyUsage |= x509.KeyUsageKeyEncipherment
	template.KeyUsage |= x509.KeyUsageKeyAgreement

	// generate private key, ca.pem
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer func() { _ = certOut.Close() }()
	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	//write private key
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err

	}
	defer func() { _ = keyOut.Close() }()

	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

// generateCert generates a new certificate signed using the provided
// certificate authority files and stores the result in the certificate
// file and key provided.  The provided host names are set to the
// appropriate certificate fields.
func generateCert(org, caFile, caKeyFile, certFile, keyFile string, bits int, hosts []string) error {
	template, err := newCertificate(org)
	if err != nil {
		return err
	}
	// client
	if len(hosts) == 1 && hosts[0] == "" {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		template.KeyUsage = x509.KeyUsageDigitalSignature
	} else { // server
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

		for _, h := range hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(caFile, caKeyFile)
	if err != nil {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, x509Cert, &priv.PublicKey, tlsCert.PrivateKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer func() { _ = certOut.Close() }()
	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = keyOut.Close() }()

	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

func newCertificate(org string) (*x509.Certificate, error) {
	now := time.Now()
	// need to set notBefore slightly in the past to account for time
	// skew in the VMs otherwise the certs sometimes are not yet valid
	notBefore := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-5, 0, 0, time.Local)
	notAfter := notBefore.Add(time.Hour * 24 * 365 * 100)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement,
		BasicConstraintsValid: true,
	}, nil
}
