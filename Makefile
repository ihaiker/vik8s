.PHONY: help chmod build cicd mkdocs cicd certs testcerts

help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

chmod:
	chmod +x ./scripts/*.sh

build: chmod ## 编译程序
	./scripts/build.sh

cicd: build ## 运行CI/CD测试
	./scripts/cicd.sh

ssh: ## CentOS使用root登录
	ssh -q root@10.24.0.10 -i .vagrant/machines/master0/virtualbox/private_key

certs: build ## 生成docker证书
	./bin/vik8s -f ./bin docker --tls.enable --hosts "tcp://{IP}:2375"

testcerts: ## 测试证书
	DOCKER_TLS_VERIFY="0" \
   	DOCKER_HOST="tcp://10.24.0.10:2375" \
	DOCKER_CERT_PATH=~/workbench/self/go/vik8s/bin/default/ \
	DOCKER_CERT_PATH=~/workbench/self/go/vik8s/bin/default/docker/certs.d/ \
	docker ps -a

clean: ## 清理
	vagrant destroy -f
	rm -rf ./bin .vagrant

mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

