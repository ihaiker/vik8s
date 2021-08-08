bin=./bin/vik8s -f ./bin/

.PHONY: help
help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: chmod
chmod:
	chmod +x ./scripts/*.sh

.PHONY: addhosts
addhosts: build
	$(bin) hosts -u vagrant -i .vagrant/machines/master01/virtualbox/private_key 10.24.0.10
	$(bin) hosts -u vagrant -i .vagrant/machines/master01/virtualbox/private_key 10.24.0.10
	$(bin) hosts -u vagrant -i .vagrant/machines/master01/virtualbox/private_key 10.24.0.10

.PHONY: build
build: chmod ## 编译程序
	./scripts/build.sh

.PHONY: cicd
cicd: build ## 运行CI/CD测试
	./scripts/cicd.sh

.PHONY: docker
docker: build ## 生成docker证书
	$(bin) docker --tls.enable --hosts "tcp://{IP}:2375"

.PHONY: etcd
etcd: addhosts
	$(bin) etcd init 10.24.0.10
	$(bin) etcd join 10.24.0.20
	$(bin) etcd join 10.24.0.21

.PHONY: etcd-clean
etcd-clean:
	$(bin) etcd reset

.PHONY: init
init: addhosts
	$(bin) init 10.24.0.10

.PHONY: mkdocs
mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

.PHONY: clean
clean: ## 清理
	vagrant destroy -f
	rm -rf ./bin .vagrant
