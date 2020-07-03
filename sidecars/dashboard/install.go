package dashboard

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/flags"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
	"strconv"
)

type Dashboard struct {
	EnableInsecureLogin bool `flag:"enable-insecure-login" help:"When enabled, Dashboard login view will also be shown when Dashboard is not served over HTTPS." def:"false"`
	/*
		nginx   : https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
		traefik : https://docs.traefik.io/v1.7/configuration/backends/kubernetes/
	*/
	InsecureHeader bool `flag:"insecure-header" help:"Add secure access control token to the header in ingress." def:"false"`

	Expose int `help:"expose dashboard server nodeport. -1: disable, 0: system allocation, >0: designated port" def:"-1"`

	Ingress string `help:"deploy dashboard ingress host name"`
	TlsKey  string `help:"dashboard tls dashboard.key path"`
	TlsCert string `help:"dashboard tls dashboard.crt path"`
}

func (d *Dashboard) Name() string {
	return "dashboard"
}

func (d *Dashboard) Description() string {
	return `Web UI (Dashboard)
more info : https://github.com/kubernetes/dashboard/blob/master/docs/user/README.md`
}

func (d *Dashboard) Flags(cmd *cobra.Command) {
	flags.Flags(cmd.Flags(), d, "")
	cmd.AddCommand(tokenPrintCmd)
}

func (d *Dashboard) Apply() {
	exposePort := d.Expose
	enableInsecureLogin := d.EnableInsecureLogin
	insecureHeader := d.InsecureHeader

	ingress := d.Ingress
	certPath := d.TlsCert
	keyPath := d.TlsKey

	master := k8s.Config.Master()

	certBase64, keyBase64 := makeDashboardCertAndKey(ingress, certPath, keyPath)
	//dashboard
	{
		data := tools.Json{"ExposePort": exposePort}
		if enableInsecureLogin {
			reduce.MustApplyAssert(master, "yaml/sidecars/dashboard/alternative.conf", data)
		} else {
			data["TLSCert"], data["TLSKey"] = certBase64, keyBase64
			reduce.MustApplyAssert(master, "yaml/sidecars/dashboard/recommended.conf", data)
		}
	}

	//dashboard access control
	token := ""
	{
		reduce.MustApplyAssert(master, "yaml/sidecars/dashboard/user.conf", tools.Json{})
		token = master.MustCmd2String(tokenPrintCmd.Long)
	}

	if ingress != "" {
		data := tools.Json{
			"Ingress": ingress, "Token": token,
			"EnableInsecureLogin": enableInsecureLogin, "InsecureHeader": insecureHeader,
			"TLSCert": certBase64, "TLSKey": keyBase64,
		}
		reduce.MustApplyAssert(master, "yaml/sidecars/dashboard/ingress.conf", data)
	}

	//show access function
	if exposePort == 0 {
		allocExposePort := master.MustCmd2String(`kubectl get -n kubernetes-dashboard service kubernetes-dashboard -o jsonpath={.spec.ports[0].nodePort}`)
		exposePort, _ = strconv.Atoi(allocExposePort)
	}

	fmt.Println(`Successful installation.`)

	scheme := "https"
	if enableInsecureLogin {
		scheme = "http"
	}

	if ingress != "" || exposePort >= 0 {
		fmt.Println("You can access the address via the URL")
		if exposePort >= 0 {
			fmt.Printf("\t%s://%s:%d\n", scheme, master.Host, exposePort)
		}
		if ingress != "" {
			fmt.Printf("\t%s://%s\n", scheme, ingress)
		}
	} else {
		proxyPort := 8443
		if enableInsecureLogin {
			proxyPort = 9090
		}
		fmt.Printf(`
To access Dashboard from your local workstation you must create a secure channel to your Kubernetes cluster. Run the following command:
$ kubectl proxy
Now access Dashboard at:
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/%s:kubernetes-dashboard:%d/proxy/.
`, scheme, proxyPort)
	}
	if enableInsecureLogin && !insecureHeader {
		fmt.Println("To make Dashboard use authorization header you simply need to pass Authorization: Bearer <token> in every request to Dashboard.")
		fmt.Println("How to access, please check the documentation help yourself：\n" +
			"   https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/README.md#authorization-header")
	}
	fmt.Println("\ntoken: ")
	fmt.Println(token)
	fmt.Println(`	
Accessing Dashboard: 
	https://github.com/kubernetes/dashboard/blob/5e86d6d405df3f85fe13938501689c663fdb9fb0/docs/user/accessing-dashboard/README.md`)
}

func (d *Dashboard) Delete(data bool) {
	master := k8s.Config.Master()
	fmt.Println(master.MustCmd2String("kubectl delete namespaces kubernetes-dashboard"))
}

func makeDashboardCertAndKey(commonName, tlsCertPath, tlsKeyPath string) (tlsCertBase64, tlsKeyBase64 string) {

	if tlsCertPath != "" && tlsKeyPath != "" {
		tlsCertBase64 = utils.Base64File(tlsCertPath)
		tlsKeyBase64 = utils.Base64File(tlsKeyPath)
		return
	}

	if commonName == "" {
		commonName = "kubernetes-dashboard"
	}

	cert, key := certs.NewCertificateAuthority(certs.NewConfig(commonName))
	cfg := certs.NewConfig(commonName)
	cfg.AltNames = *certs.GetAltNames([]string{commonName}, commonName)
	cfg.Usages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	cert, key = certs.NewCertAndKey(cert, key, cfg)

	tlsCertBase64 = base64.StdEncoding.EncodeToString(certs.EncodeCertPEM(cert))
	tlsKeyBase64 = base64.StdEncoding.EncodeToString(certs.EncodePrivateKeyPEM(key))
	return
}
