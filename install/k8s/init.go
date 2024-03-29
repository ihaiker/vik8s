package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func InitCluster(configure *config.Configuration, node *ssh.Node) *ssh.Node {
	node.Logger("init kubernetes cluster %s", node.Host)

	configure.K8S.Repo = repo.KubeletImage(configure.K8S.Repo)
	if configure.K8S.ApiServerVIP == "" {
		configure.K8S.ApiServerVIP = tools.GetVip(configure.K8S.SvcCIDR, tools.Vik8sApiServer)
	}

	bases.Check(node)
	bases.InstallTimeServices(node, configure.K8S.Timezone, configure.K8S.NTPServices...)
	bases.InstallJQTools(node)
	cri.Install(configure, node)

	installKubernetesSoftware(configure, node)

	node.Logger("init cluster")
	{
		setHosts(node, node.Host, configure.K8S.ApiServer)
		setHosts(node, node.Host, node.Hostname)
		if node.Hostname != node.Facts.Hostname {
			setHosts(node, node.Host, node.Facts.Hostname)
		}
		makeKubernetesCerts(configure, node)
		makeJoinControlPlaneConfigFiles(configure, node)
		initKubernetes(configure, node)
		copyKubeletAdminConfig(node)
		applyApiServerEndpoint(configure, node)
	}
	return node
}

func setHosts(node *ssh.Node, ip, domain string) {
	node.Logger("set host %s => %s", ip, domain)
	hostsContent, err := node.Sudo().CmdString("cat /etc/hosts")
	utils.Panic(err, "fetch hosts list")

	hosts := strings.Split(hostsContent, "\n")
	findLine := -1
	editFn := 0 //0: append, 1: edit
	pattern := regexp.MustCompile("\\s+")
	for line, ipAndDomainsStr := range hosts {
		//trim all space
		ipAndDomainsStr = strings.TrimRight(ipAndDomainsStr, "")
		if pattern.ReplaceAllString(ipAndDomainsStr, " ") == "" ||
			strings.HasPrefix(ipAndDomainsStr, "#") {
			continue
		}

		ipAndDomains := pattern.Split(ipAndDomainsStr, -1)
		if idx := utils.Search(ipAndDomains[1:], domain); idx != -1 {
			if ip == ipAndDomains[0] { // no need to modify it
				return
			}
			if len(ipAndDomains) == 2 {
				findLine = line + 1
				editFn = 1
			} else {
				ipAndDomains[idx+1] = ""
				err = node.Sudo().Cmd(fmt.Sprintf("sed -i '%dc%s' /etc/hosts", line+1, strings.Join(ipAndDomains, " ")))
				utils.Panic(err, "set /etc/hosts")
			}
			break
		}
	}

	if editFn == 0 { //0: append, 1: edit
		err = node.Sudo().Cmd(fmt.Sprintf("sed -i '$ a%s %s' /etc/hosts", ip, domain))
		utils.Panic(err, "set /etc/hosts")
	} else {
		err = node.Sudo().Cmd(fmt.Sprintf("sed -i '%dc%s %s' /etc/hosts", findLine, ip, domain))
		utils.Panic(err, "set /etc/hosts")
	}
}

func bugfixImages(master, node *ssh.Node, remote string) {
	images, err := master.Sudo().CmdString(fmt.Sprintf("kubeadm config images list --config=%s", remote))
	utils.Panic(err, "list kubernetes images")

	//fix: alicloud image repo. since kubeadm@v1.21.+
	tags := map[string]string{
		"registry.aliyuncs.com/google_containers/coredns:1.8.0": "registry.aliyuncs.com/google_containers/coredns:v1.8.0",
	}
	for imageSource, imageDest := range tags {
		if strings.Contains(images, imageDest) {
			node.Logger("bugfix: image %s not found", imageDest)

			err = node.Sudo().CmdOutput("docker pull "+imageSource, os.Stdout)
			utils.Panic(err, "pull docker images")

			err = node.Sudo().CmdOutput("docker tag "+imageSource+" "+imageDest, os.Stdout)
			utils.Panic(err, "tag docker images")
		}
	}
}

func initKubernetes(configure *config.Configuration, node *ssh.Node) {
	remote := scpKubeConfig(configure, node)
	bugfixImages(node, node, remote)
	err := node.Sudo().CmdOutput(fmt.Sprintf("kubeadm init --config=%s --upload-certs", remote), os.Stdout)
	utils.Panic(err, "kubeadm init")
}

func scpKubeConfig(configure *config.Configuration, node *ssh.Node) string {
	var kubeadmConfigBytes []byte
	var err error

	tmpData := templateDate(configure, node)

	if configure.K8S.KubeadmConfig != "" {
		configBytes, err := ioutil.ReadFile(configure.K8S.KubeadmConfig)
		utils.Panic(err, "read kubeadm-config file %s", configure.K8S.KubeadmConfig)
		kubeadmConfigBytes, err = tools.Template(string(configBytes), tmpData)
		utils.Panic(err, "parse user kubeadm config error: %s", configure.K8S.KubeadmConfig)
	} else {
		kubeadmConfigBytes, err = tools.Assert("yaml/kubeadm-config.yaml", tmpData)
		utils.Panic(err, "parse kubeadm config error")
	}

	remote := node.Vik8s("apply/kubeadm.yaml")
	err = node.Sudo().ScpContent(kubeadmConfigBytes, remote)
	utils.Panic(err, "scp kubeadm-config file")
	return remote
}

func copyKubeletAdminConfig(node *ssh.Node) {
	err := node.Cmd("mkdir -p $HOME/.kube")
	utils.Panic(err, "copy kubectl config")

	err = node.Sudo().Cmd("cp -f /etc/kubernetes/admin.conf $HOME/.kube/config")
	utils.Panic(err, "copy kubectl config")

	err = node.Sudo().Cmd("chown -R $(id -u):$(id -g) $HOME/.kube")
	utils.Panic(err, "change vik8s config dir")
}

func applyApiServerEndpoint(configure *config.Configuration, node *ssh.Node) {
	name := "yaml/vik8s-api-server.conf"
	data := templateDate(configure, node)
	//tools.MustScpAndApplyAssert(node, name, data)
	err := reduce.ApplyAssert(node, name, data)
	utils.Panic(err, "apply vik8s-api-server")
}

func templateDate(configure *config.Configuration, node *ssh.Node) paths.Json {
	masters := append(configure.Hosts.MustGets(configure.K8S.Masters), node)
	nodes := configure.Hosts.MustGets(configure.K8S.Nodes)
	data := paths.Json{
		"Masters": masters, "Workers": nodes,
		"nodes": append(masters, nodes...), "Kubeadm": configure.K8S,
	}
	if configure.IsExternalETCD() {
		endpoints := make([]string, 0)
		if configure.ExternalETCD != nil {
			endpoints = configure.ExternalETCD.Endpoints
		} else {
			for _, node := range configure.ETCD.Nodes {
				endpoints = append(endpoints, fmt.Sprintf("https://%s:2379", node))
			}
		}
		data["Etcd"] = paths.Json{
			"External": true, "Endpoints": endpoints,
		}
	} else {
		data["Etcd"] = paths.Json{"External": false}
	}
	data["Kubeadm"] = configure.K8S
	return data
}
