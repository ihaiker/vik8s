.PHONY: help
help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: chmod
chmod:
	chmod +x ./scripts/*.sh

.PHONY: build
build: chmod ## 编译程序
	./scripts/build.sh

.PHONY: cicd
cicd: build ## 运行CI/CD测试
	./scripts/cicd.sh

.PHONY: ssh
ssh: ## CentOS使用root登录
	ssh -q root@10.24.0.10 -i .vagrant/machines/master0/virtualbox/private_key

.PHONY: ssh-slave20
ssh-slave20: ## CentOS使用slave20登录
	ssh -q root@10.24.0.20 -i .vagrant/machines/slave20/virtualbox/private_key

.PHONY: ssh-slave21
ssh-slave21: ## CentOS使用slave21登录
	ssh -q root@10.24.0.21 -i .vagrant/machines/slave21/virtualbox/private_key

.PHONY: docker
docker: build ## 生成docker证书
	./bin/vik8s -f ./bin docker --tls.enable --hosts "tcp://{IP}:2375"

.PHONY: etcd
etcd: build
	./bin/vik8s -f ./bin/ etcd init 10.24.0.10
	./bin/vik8s -f ./bin/ etcd join 10.24.0.20
	./bin/vik8s -f ./bin/ etcd reset

.PHONY: mkdocs
mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

.PHONY: clean
clean: ## 清理
	vagrant destroy -f
	rm -rf ./bin .vagrant
