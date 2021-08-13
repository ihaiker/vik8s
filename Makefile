.PHONY: help chmod build vagrant esxi cicd mkdocs clean

help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

chmod: ## 脚本赋权
	chmod +x ./scripts/*.sh

build: chmod ## 编译程序
	./scripts/build.sh

vagrant: build ## 启动虚拟机
	./scripts/vagrant.sh up
	./scripts/hosts.sh

cicd: vagrant ## 运行CI/CD测试
	./scripts/cicd.sh

mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

clean: ## 清理
	./scripts/vagrant.sh destroy -f
	rm -rf ./bin .vagrant
