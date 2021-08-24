package dashboard

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
	"strconv"
)

type dashboard struct {
	EnableInsecureLogin bool `flag:"enable-insecure-login" help:"When enabled, dashboard login view will also be shown when dashboard is not served over HTTPS." def:"false"`
	/*
		nginx   : https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
	*/
	InsecureHeader bool `flag:"insecure-header" help:"Add secure access control token to the header in ingress." def:"false"`

	Expose int `help:"expose dashboard server nodeport. -1: disable, 0: system allocation, >0: designated port" def:"-1"`

	Ingress string `help:"deploy dashboard ingress host name"`
	TlsKey  string `help:"dashboard tls dashboard.key path"`
	TlsCert string `help:"dashboard tls dashboard.crt path"`
}

func New() *dashboard {
	return &dashboard{
		EnableInsecureLogin: false,
		InsecureHeader:      false,
		Expose:              -1,
	}
}

func (d *dashboard) Name() string {
	return "dashboard"
}

func (d *dashboard) Description() string {
	return `Web UI (dashboard)
more info : https://github.com/kubernetes/dashboard/blob/master/docs/user/README.md`
}

func (d *dashboard) Flags(cmd *cobra.Command) {
	err := cobrax.Flags(cmd, d, "", "")
	utils.Panic(err, "set dashboard flags")
	cmd.AddCommand(tokenPrintCmd)
	tokenPrintCmd.PreRunE = cmd.PreRunE
}

func (d *dashboard) Apply() {
	exposePort := d.Expose
	enableInsecureLogin := d.EnableInsecureLogin
	insecureHeader := d.InsecureHeader

	ingress := d.Ingress
	certPath := d.TlsCert
	keyPath := d.TlsKey

	master := hosts.Get(config.K8S().Masters[0])

	certBase64, keyBase64 := makeDashboardCertAndKey(ingress, certPath, keyPath)
	//dashboard
	{
		data := paths.Json{"ExposePort": exposePort}
		if enableInsecureLogin {
			err := reduce.ApplyAssert(master, "yaml/sidecars/dashboard/alternative.conf", data)
			utils.Panic(err, "apply alternative kubernetes dashboard ")
		} else {
			data["TLSCert"], data["TLSKey"] = certBase64, keyBase64
			err := reduce.ApplyAssert(master, "yaml/sidecars/dashboard/recommended.conf", data)
			utils.Panic(err, "apply recommended kubernetes dashboard ")
		}
	}

	//dashboard access control
	token := ""
	{
		err := reduce.ApplyAssert(master, "yaml/sidecars/dashboard/user.conf", paths.Json{})
		utils.Panic(err, "apply kubernetes dashboard user")
		token, err = master.Sudo().CmdString(tokenPrintCmd.Long)
		utils.Panic(err, "apply kubernetes dashboard user")
	}

	if ingress != "" {
		data := paths.Json{
			"Ingress": ingress, "Token": token,
			"EnableInsecureLogin": enableInsecureLogin, "InsecureHeader": insecureHeader,
			"TLSCert": certBase64, "TLSKey": keyBase64,
		}
		err := reduce.ApplyAssert(master, "yaml/sidecars/dashboard/ingress.conf", data)
		utils.Panic(err, "apply kubernetes dashboard ingress")
	}

	//show access function
	if exposePort == 0 {
		allocExposePort, err := master.Sudo().CmdString(`kubectl get -n kubernetes-dashboard service kubernetes-dashboard -o jsonpath={.spec.ports[0].nodePort}`)
		utils.Panic(err, "get dashboard nodePort")
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
To access dashboard from your local workstation you must create a secure channel to your Kubernetes cluster. Run the following command:
$ kubectl proxy
Now access dashboard at:
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/%s:kubernetes-dashboard:%d/proxy/.
`, scheme, proxyPort)
	}
	if enableInsecureLogin && !insecureHeader {
		fmt.Println("To make dashboard use authorization header you simply need to pass Authorization: Bearer <token> in every request to dashboard.")
		fmt.Println("How to access, please check the documentation help yourself：\n" +
			"   https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/README.md#authorization-header")
	}
	fmt.Println("\ntoken: ")
	fmt.Println(token)
	fmt.Println(`	
Accessing dashboard: 
	https://github.com/kubernetes/dashboard/blob/5e86d6d405df3f85fe13938501689c663fdb9fb0/docs/user/accessing-dashboard/README.md`)
}

func (d *dashboard) Delete(data bool) {
	master := hosts.Get(config.K8S().Masters[0])
	err := master.Sudo().CmdPrefixStdout("kubectl delete namespaces kubernetes-dashboard")
	utils.Panic(err, "remove kubernests cluster namespace kubernetes-dashboard")
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
