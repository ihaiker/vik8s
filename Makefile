.PHONY: help build init clean reset generate release


Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)
BuildDate=$(shell date +"%F %T")
GitCommit=$(shell git rev-parse --short HEAD)
debug=-w -s
param=-X main.VERSION=${Version} -X main.GITLOG_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'


build: ## 编译程序
	go generate
	go build -o /usr/local/bin/vik8s main.go
	vik8s completion > /usr/local/etc/bash_completion.d/vik8s

init: build ## 初始化一个节点
	vik8s init -m 172.16.100.10 --cni calico --cni-calico-etcd

joinm: build ## 加入新节点
	vik8s join -m 172.16.100.14 -m 172.16.100.15 -n 172.16.100.11-172.16.100.13

clean: build ## 清除全部节点（重要，会删除文件夹）
	vik8s clean --force

reset: build ## 清楚全部节点
	vik8s reset all

all: reset
	vik8s init -m 172.16.100.10 --cni calico --cni-calico-etcd
	vik8s join -m 172.16.100.14 -m 172.16.100.15 -n 172.16.100.11-172.16.100.13
	vik8s ingress traefik --host-network=true --ui-ingress=traefik.vik8s.io --ui-passwd=haiker --node.selector=kubernetes.io/hostname=vm11
	vik8s sidecars dashboard --expose=31715 --enable-insecure-login --insecure-header --ingress=dashboard.vik8s.io

etcd-init: build ## etcd集群初始化
	vik8s etcd init 172.16.100.11-172.16.100.13

etcd-join: build ## etcd加入新节点
	vik8s etcd join 172.16.100.14

etcd-reset: build ## etcd集群重置
	vik8s etcd reset 172.16.100.11-172.16.100.13

help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

ingress-traefik: build
	vik8s ingress traefik del
	vik8s ingress traefik --host-network=true --ui-ingress=traefik.vik8s.io --ui-passwd=haiker --node.selector=kubernetes.io/hostname=vm11

ingress-nginx: build
	vik8s ingress nginx --host-network=true --node.selector=kubernetes.io/hostname=vm11

dashboard: build
	vik8s sidecars dashboard uninstall
	vik8s sidecars dashboard --expose=31715 --enable-insecure-login --insecure-header --ingress=dashboard.vik8s.io

glusterfs: build
	vik8s storage glusterfs uninstall --data
	vik8s -- rm -rf /etc/glusterfs/* /var/lib/glusterd/* /var/lib/heketi/* /data/glusterfs/volumes
	vik8s storage glusterfs --heketi.enable=false --nodes=vm11
	vik8s storage glusterfs volume chain
	vik8s storage glusterfs peer --volume=chain vm12 vm13

scp:
	scp bin/glusterfs.yaml vm10:/root
	ssh vm10 kubectl apply -f /root/glusterfs.yaml

release:
	go generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${debug} ${param}" -o bin/vik8s-linux-amd64-${Version} main.go
	tar -czvf bin/vik8s-linux-amd64-${Version}.tar.gz bin/vik8s-linux-amd64-${Version}

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "${debug} ${param}" -o bin/vik8s-darwin-amd64-${Version} main.go
	tar -czvf bin/vik8s-darwin-amd64-${Version}.tar.gz bin/vik8s-darwin-amd64-${Version}

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${debug} ${param}" -o bin/vik8s-windows-amd64-${Version}.exe main.go
	zip bin/vik8s-windows-amd64-${Version}.zip bin/vik8s-windows-amd64-${Version}.exe