# QuickStart

## 安装

### 1、下载二进制文件

[Github https://github.com/ihaiker/vik8s](https://github.com/ihaiker/ngx/releases/latest)

下载和平台相应的二进制包


### 2、编译安装

```shell
git clone https://github.com/ihaiker/vik8s
./scripts/build.sh
cp bin/vik8s /usr/local/bin
```



### 检验版本信息

执行 vik8s 将看到一下信息。

```shell
very easy install k8s。Build: 2021-11-20 13:39:10, Go: go1.15.6, GitLog: 0d88c1b2296c9cd79d19d0108e99063d68217c65

Usage:
  vik8s [command]

Available Commands:
  bash        Run commands uniformly in the cluster
  cni         define kubernetes network interface
  completion  generates completion scripts
  config      Show yaml file used by vik8s deployment cluster
  cri         defined kubernetes container runtime interface
  etcd        Install ETCD cluster
  help        Help about any command
  hosts       Add Management Host
  ingress     install kubernetes ingress controller
  init        Initialize the kubernetes cluster
  join        join to k8s
  reduce      Simplify kubernetes configuration file
  reset       reset kubernetes cluster node

Flags:
      --china           Whether domestic network (default true)
  -c, --cloud string    Multi-kubernetes cluster selection (default "default")
  -f, --config string   The folder where the configuration file is located (default "~/.vik8s")
  -h, --help            help for vik8s
  -v, --version         version for vik8s

Use "vik8s [command] --help" for more information about a command.
```

## 主机准备

| 主机名称 | IP地址        |
| -------- | ------------- |
| master1  | 172.16.100.10 |
| master2  | 172.16.100.11 |
| master3  | 172.16.100.12 |
| node1    | 172.16.100.13 |
| node2    | 172.16.100.14 |
| node3    | 172.16.100.15 |

<span style="color:red;">注意：所有主机采用配置root免密登录，并且端口为22。</span> 如果您的主机访问方式存在差异，那您可以查看[主机访问方式管理](./cmds/hosts/index.md)



## 安装

1、master节点安装

```shell
$ vik8s init 172.16.100.10 # 初始化集群
$ vik8s join --master 172.16.100.11 172.16.100.12 #添加两个控制节点
```

2、slave节点安装

```shell
vik8s join 172.16.100.13-172.16.100.15
```

3、网络插件安装

```shell
vik8s cni flannel
```

更多详细信息查看[Kubernetes集群初始化教程](./cmds/k8s/index.md)和[网络插件安装](./cmds/k8s/cni.md)

