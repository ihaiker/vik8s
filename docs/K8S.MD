
# kubernetes集群安装

## 环境准备

|  主机名称  | IP地址 |
| -------- | ------------ |
| master1 | 172.16.100.10 |
| master2 | 172.16.100.11 |
| master3 | 172.16.100.12 |
| node1 | 172.16.100.13 |
| node2 | 172.16.100.14 |
| node3 | 172.16.100.15 |


## 控制节点安装

1、master1 安装
```shell script
vik8s init --master 172.16.100.10
```
2、master 2/3 安装
```shell script
vik8s join --master 172.16.100.11 --node 172.16.100.12
```
3、node1-3安装
```shell script
vik8s join --node 172.16.100.13-172.16.100.15 #连续IP使用
```
注：上面的命令所有节点都是22端口并且采用证书登录。并且证书放在$HOME/.ssh/id_rsa下，且网络插件采用flannel并自动部署。

## 使用非SSH免密登录方式

使用密码方式初始化 添加命令参数 `--ssh-passwd=password` 即可, 更改SSH端口 添加命令参数 `--ssh-port=123`。当然了本程序为了降低开发难度并未实现非ROOT用户方式。

极端情况：在本程序开发过程中为了保证所有场景下均方便（不同主机SSH的端口和密码或者免密文件不一致），提供了一种万能的方式，就是所有主机可以单独设置连接方式的验证方式和密码。

> 第一种：在`--master`和`--node`参数上直接设置 格式：`root:paswword@ip-ip:port`

例如: `--master username:password@10.24.0.10:22` 或者连续IP `--master username:password@10.24.0.13-10.24.0.33:22`。
注意不管`username`设置成什么，本程序都会改为root，暂时不支持其他用户，这样做是为了以后预留功能。
    
> 第二种：使用hosts命令先行管理所有主机。

通过命令 `vik8s hosts --passwd=password --port=1232 172.16.100.10 172.16.100.11 172.16.100.12` 
或者 `vik8s hosts --pk=~/.ssh/id_rsa --port=22 172.16.100.13-172.16.100.15` 添加到host管理器中。
之后我们在使用添加命令的时候就可以直接使用`hostname`去代替IP了。我们先行添加host管理后，初始化命令就可以改为`vik8s init --master master1`和`vik8s join -m master2 -n node1`

有兴趣的朋友可以[查看源码][1]

注意：无论你是否使用第二种方式，在集群加入后会自动添加主机的信息到host管理中，因此整个程序中在某些命令中的节点选择方式均是采用的这种方式。

## 命令详解

| 参数                    | 解释                                                         |
| ----------------------- | ------------------------------------------------------------ |
| --k8s-version string | 安装版本, 支持 1.17.+ (default "1.18.2") |
|  |  |
| -P, --ssh-port int      | SSH默认端口 （默认：22）                       |
| --ssh-pk string         | SSH私钥路径 (default "$HOME/.ssh/id_rsa")       |
| -p, --ssh-passwd string | SSH密码password，如果提供了密码将使用密码     |
|  |  |
| -m, --master strings    | k8s控制节点。 可以是单独IP，也可以是连续IP，例如:172.16.100.10 或者172.16.100.10-172.16.100.20 |
| -n, --node strings      | k8s工作节点。同：--master |
|                         |                                                              |
| --docker-version string             | 如果没有安装docker将自动安装docker的版本号 (default "19.03.8") |
| --docker-registry string            | 设置 docker 私服地址，如果设置了--docker-daemon参数将忽略此参数 |
| --docker-check-version              | Mandatory check DOCKER version number will upgrade if inconsistent |
| --docker-daemon string | docker的配置文件 /etc/docker/daemon.json |
|  |  |
| --kubeadm-config string              | kubeadm配置文件.  此文件将被使用在 初始化 kubeadm --config 命令上。 你可以通过 `vik8s config yaml/kubeadm-config.yaml` 查看默认配置。 |
| --apiserver string                   | 指定指定HA高可用控制节点apiserver的dns名称。 see kubeadm  --control-plane-endpoint (默认 "vik8s-api-server") |
| --apiserver-cert-extra-sans strings  | 详细查看  kubeadm init --apiserver-cert-extra-sans |
|  |  |
| --interface string                   | 集群使用网卡名称 (default  "eth.*\|en.*\|em.* ") |
| --pod-cidr string                    | 指定POD网络范围  (default "100.64.0.0/24") |
| --svc-cidr string                    | 指定Services网络范围 (default "10.96.0.0/12") |
| --repo string                        | kubernetes集群镜像下载地址.  默认从 k8s.gcr.io 和 registry.aliyuncs.com/google_containers 中选择 |
|                                      |  |
| --certs-validity duration            | Certificate validity time (default 87648h0m0s) |
| --etcd                               | 使用外部ETCD集群. 如果您使用了 `vik8s etcd init` 命令安装了etcd |
| --etcd-endpoints strings             | 外部ETCD集群地址:  172.16.100.10:2379 |
| --etcd-ca string                     | 外部ETCD集群ca证书 |
| --etcd-apiserver-key-file string     | 外部ETCD集群apiserver证书 |
| --etcd-apiserver-cert-file string    |  |
|  |  |
| --cni string                         | 使用的网络插件. 系统支持: ignore,flannel,calico,customer (default "flannel")<br />ignore:        忽略网络插件安装，如果您想自己安装其他网络插件<br />customer:  使用自定义的网络查看，如果设置了就需要提供自定义插件的部署路径。 |
|  |  |
| --cni-flannel-version string         | flannel网络插件版本号 (default "0.12.0") |
| --cni-flannel-limits-cpu string      | Container Cup Limit (default "100m") |
| --cni-flannel-limits-memory string    | Container Memory Limit (default "50Mi") |
| --cni-flannel-repo string             | 镜像下载地址 默认 从 quay.mirrors.ustc.edu.cn （国内） 或者 quay.io 下载 |
|  |  |
| --cni-calico-version string           | caolico 网络插件版本号 (default "3.14.0") |
| --cni-calico-ipip                     | Enable IPIP (default true) |
| --cni-calico-mtu int                  |  |
| --cni-calico-repo string              | Choose a container registry to pull control plane images from |
| --cni-calico-typha                    | 是否启用 Typha 方式存储 calico数据。 |
| --cni-calico-typha-prometheus         | 是否启用 prometheus metrics. |
| --cni-calico-typha-replicas int                  | typea 部署 个数。see Deployment 'calico-typha' at https://docs.projectcalico.org/manifests/calico-typha.yaml (default 1) |
| --cni-calico-etcd                     | 是否使用ETCD存储calico数据.<br />如果启用了etcd存储数据，但是并未提供`--cni-calico-etcd-endpoints` 参数。系统将从下面两个方面查找etcd<br />1. 使用 `--etcd`提供的etcd集群<br />2. 如果未提供`--etcd` 系统将使用控制节点提供的etcd . |
| --cni-calico-etcd-tls                 | etcd是否开始tls  (default true) |
| --cni-calico-etcd-endpoints strings   | 172.16.100.10:2379 |
| --cni-calico-etcd-ca string           |  |
| --cni-calico-etcd-key string          |  |
| --cni-calico-etcd-cert string         |  |
|  |  |
| --cni-customer-url string             | 用户自定网络插件地址 |
| --cni-customer-file string            | 用户自定义网络插件文件 |
|  |  |
| --timezone string                    | 服务器时区 (default "Asia/Shanghai") |
| --ntp-services strings               | 时间服务器 (default [ntp1.aliyun.com,ntp2.aliyun.com,ntp3.aliyun.com]) |



[1]: "https://github.com/ihaiker/aginx/blob/master/nginx/analysis.go" "host 实现方式"