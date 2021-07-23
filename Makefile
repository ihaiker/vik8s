.PHONY: help chmod build cicd mkdocs

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

clean: ## 清理
	vagrant destroy -f
	rm -rf ./bin .vagrant

mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

