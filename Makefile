.PHONY: help chmod build vagrant cicd mkdocs clean test

help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

chmod: ## 脚本赋权
	chmod +x ./scripts/*.sh

test: ## 测试GO代码
	go test ./... -failfast -cover

build: chmod ## 编译程序
	./scripts/build.sh

vagrant: ## 启动虚拟机
	vagrant up

cicd: ## 运行CI/CD测试¸
	./scripts/cicd.sh

mkdocs: ## 构建文档
	docker run --rm -v `pwd`/:/build xhaiker/mkdocs:latest build -c

clean: ## 清理
	vagrant destroy -f
	rm -rf ./bin .vagrant
