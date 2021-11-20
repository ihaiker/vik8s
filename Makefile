.PHONY: help chmod build vagrant esxi cicd mkdocs clean test release

help: ## 帮助信息
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

chmod: ## 脚本赋权
	chmod +x ./scripts/*.sh

test: ## 测试GO代码
	go test ./... -failfast -cover

build: chmod ## 编译程序
	./scripts/build.sh

vagrant: build ## 启动虚拟机
	vagrant up
	./scripts/hosts.sh

cicd: ## 运行CI/CD测试
	./scripts/cicd.sh

mkdocs: ## 构建文档
	docker-compose -f ./scripts/docker-compose.yml run --rm mkdocs build -c

release: ## 构建Release
	$PWD=`pwd`
	docker run --rm --privileged \
      -v ${PWD}:/go/src/github.com/ihaiker/vik8s \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -w /go/src/github.com/ihaiker/vik8s \
	  -e GITHUB_TOKEN=123456789123456789 \
	  -e GOPROXY=https://goproxy.cn,direct \
      goreleaser/goreleaser release --rm-dist --release-header-tmpl=./docs/releases/v0.5.0.md

clean: ## 清理
	./scripts/vagrant.sh destroy -f
	rm -rf ./bin .vagrant
