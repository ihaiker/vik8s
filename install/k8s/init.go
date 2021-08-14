package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	yamls "github.com/ihaiker/vik8s/yaml"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func InitCluster(node *ssh.Node) *ssh.Node {
	node.Logger("init kubernetes cluster %s", node.Host)

	config.Config.K8S.Repo = repo.KubeletImage(config.K8S().Repo)
	if config.Config.K8S.ApiServerVIP == "" {
		config.Config.K8S.ApiServerVIP = tools.GetVip(config.K8S().SvcCIDR, tools.Vik8sApiServer)
	}

	bases.Check(node)
	bases.InstallTimeServices(node, config.K8S().Timezone, config.K8S().NTPServices...)
	cri.Install(node)

	installKubernetes(node)

	node.Logger("init cluster")
	{
		setHosts(node, node.Host, config.K8S().ApiServer)
		setHosts(node, node.Host, node.Hostname)
		if node.Hostname != node.Facts.Hostname {
			setHosts(node, node.Host, node.Facts.Hostname)
		}
		makeKubernetesCerts(node)
		makeJoinControlPlaneConfigFiles(node)
		initKubernetes(node)
		copyKubeletConfig(node)
		applyApiServerEndpoint(node)
	}
	config.K8S().JoinNode(true, node.Host)
	return node
}

func setHosts(node *ssh.Node, ip, domain string) {
	node.Logger("set host %s => %s", ip, domain)
	hostsContent, err := node.SudoCmdString("cat /etc/hosts")
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
				err = node.SudoCmd(fmt.Sprintf("sed -i '%dc%s' /etc/hosts", line+1, strings.Join(ipAndDomains, " ")))
				utils.Panic(err, "set /etc/hosts")
			}
			break
		}
	}

	if editFn == 0 { //0: append, 1: edit
		err = node.SudoCmd(fmt.Sprintf("sed -i '$ a%s %s' /etc/hosts", ip, domain))
		utils.Panic(err, "set /etc/hosts")
	} else {
		err = node.SudoCmd(fmt.Sprintf("sed -i '%dc%s %s' /etc/hosts", findLine, ip, domain))
		utils.Panic(err, "set /etc/hosts")
	}
}

func bugfixImages(master, node *ssh.Node, remote string) {
	images, err := master.SudoCmdString(fmt.Sprintf("kubeadm config images list --config=%s", remote))
	utils.Panic(err, "list kubernetes images")

	//fix: alicloud image repo. since kubeadm@v1.21.+
	tags := map[string]string{
		"registry.aliyuncs.com/google_containers/coredns:1.8.0": "registry.aliyuncs.com/google_containers/coredns:v1.8.0",
	}
	for imageSource, imageDest := range tags {
		if strings.Contains(images, imageDest) {
			node.Logger("bugfix: image %s not found", imageDest)

			err = node.SudoCmdOutput("docker pull "+imageSource, os.Stdout)
			utils.Panic(err, "pull docker images")

			err = node.SudoCmdOutput("docker tag "+imageSource+" "+imageDest, os.Stdout)
			utils.Panic(err, "tag docker images")
		}
	}
}

func initKubernetes(node *ssh.Node) {
	remote := scpKubeConfig(node)
	bugfixImages(node, node, remote)
	err := node.SudoCmdOutput(fmt.Sprintf("kubeadm init --config=%s --upload-certs", remote), os.Stdout)
	utils.Panic(err, "kubeadm init")
}

func scpKubeConfig(node *ssh.Node) string {
	kubeadmConfigPath := string(yamls.MustAsset("yaml/kubeadm-config.yaml"))

	if config.K8S().KubeadmConfig != "" {
		configBytes, err := ioutil.ReadFile(config.K8S().KubeadmConfig)
		utils.Panic(err, "read kubeadm-config file %s", config.K8S().KubeadmConfig)
		kubeadmConfigPath = string(configBytes)
	}

	remote := node.Vik8s("apply/kubeadm.yaml")
	node.Logger("scp kubeadm.yaml %s", remote)

	data := templateDate(node)
	kubeadmConfig := tools.Template(kubeadmConfigPath, data)
	err := node.ScpContent(kubeadmConfig.Bytes(), remote)
	utils.Panic(err, "scp kubeadm-config file")
	return remote
}

func copyKubeletConfig(node *ssh.Node) {
	err := node.Cmd2("mkdir -p $HOME/.kube")
	utils.Panic(err, "copy kubectl config")

	err = node.SudoCmd("cp -n /etc/kubernetes/admin.conf $HOME/.kube/config")
	utils.Panic(err, "copy kubectl config")

	err = node.SudoCmd("chown $(id -u):$(id -g) $HOME/.kube/config")
	utils.Panic(err, "change kubectl config owner")
}

func applyApiServerEndpoint(node *ssh.Node) {
	name := "yaml/vik8s-api-server.conf"
	data := templateDate(node)
	//tools.MustScpAndApplyAssert(node, name, data)
	reduce.MustApplyAssert(node, name, data)
}

func templateDate(node *ssh.Node) paths.Json {
	masters := append(hosts.Gets(config.K8S().Masters), node)
	nodes := hosts.Gets(config.K8S().Nodes)
	data := paths.Json{
		"Masters": masters, "Workers": nodes,
		"Nodes": append(masters, nodes...), "Kubeadm": config.K8S(),
	}
	if config.ExternalETCD() {
		data["Etcd"] = paths.Json{
			"External": true, "Nodes": config.Etcd().Nodes,
		}
	} else {
		data["Etcd"] = paths.Json{"External": false}
	}
	data["Kubeadm"] = config.K8S()
	return data
}
