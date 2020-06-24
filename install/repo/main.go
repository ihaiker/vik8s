package repo

import (
	"github.com/ihaiker/vik8s/install/tools"
	"strings"
)

func Etcdadm() string {
	if tools.China {
		return "https://gitee.com/ihaiker/etcdadm"
	} else {
		return "https://github.com/kubernetes-sigs/etcdadm"
	}
}

func Containerd() string {
	if tools.China {
		return "https://mirrors.aliyun.com/docker-ce/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.`uname -p`.rpm"
	} else {
		return "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.`uname -p`.rpm"
	}
}

func Docker() string {
	if tools.China {
		return "https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo"
	} else {
		return "https://download.docker.com/linux/centos/docker-ce.repo"
	}
}

func Kubernetes() string {
	if tools.China {
		return `[kubernetes]
baseurl = https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled = 1
gpgcheck = 1
gpgkey = https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg
        https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
name = Ali Kubernetes Repo
repo_gpgcheck = 1
`
	} else {
		return `[kubernetes]
baseurl = https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled = 1
gpgcheck = 1
gpgkey = https://packages.cloud.google.com/yum/doc/yum-key.gpg
		https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
name = Ali Kubernetes Repo
repo_gpgcheck = 1
`
	}
}

func KubeletImage() string {
	if tools.China {
		return "registry.aliyuncs.com/google_containers"
	} else {
		return "k8s.gcr.io"
	}
}

func Ceph() string {
	if tools.China {
		return "http://mirrors.aliyun.com/ceph"
	} else {
		return "https://download.ceph.com"
	}
}

func QuayIO(repo string) string {
	if repo != "" {
		return repo
	}

	if tools.China {
		return "quay.mirrors.ustc.edu.cn"
	} else {
		return "quay.io"
	}
}

func Suffix(repo string) string {
	if repo != "" && !strings.HasSuffix(repo, "/") {
		repo = repo + "/"
	}
	if strings.HasPrefix(repo, "http://") {
		repo = repo[7:]
	}
	if strings.HasPrefix(repo, "https://") {
		repo = repo[8:]
	}
	return repo
}
