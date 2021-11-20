# Kubernetes 集群初始化

## 前言：

开始本章节之前，请确定您已经查看了 [快速开始章节](../../quickstart.md)

## 命令详解

由于本程序设计构建的，你只需要以下命令

```shell
$ vik8s init 172.16.100.10 # 初始化集群
$ vik8s join --master 172.16.100.11 172.16.100.12 #添加两个控制节点
# slave节点安装
$ vik8s join 172.16.100.13-172.16.100.15
$ vik8s cni flannel #网络插件安装
```

即可简单初始化一个kubernetes集群。那么下面我们就介绍一下出初始化集群的一些参数内容：

| 参数 | 默认值 | 描述 |
| ---- | ------ | ---- |
|--version string                      |v1.21.3| 指定Kubernetes集群版本号 |
|  --kubeadm-config             || kubeadm 配置文件路径，如果您对集群有锁扩展可以设置此文件. 有关此配置文件更多信息查看`kubeadm --config` |
|  --api-server                 ||   Specify a stable IP address or DNS name for the control plane. see kubeadm --control-plane-endpoint (env: VIK8S_K8S_API_SERVER)  (default "api-vik8s-io") |
| --api-server-cert-extra-sans || kubernetes服务API证书附加sans |
|  --repo string                       || image地址获取的地址，默认从 k8s.gcr.io（国外） 和 registry.aliyuncs.com/google_containers（国内） |
|  --interface                  |`eth*`,`en*`,`em*`| 指定集群使用的网卡 |
|  --pod-cidr                   |100.64.0.0/24| 指定pod地址范围 |
|  --svc-cidr                   |10.96.0.0/12| 指定service地址范围 |
|  --certs-validity           |44y| 指定证书有效时间 |
|  --timezone                   |Asia/Shanghai|     |
|  --ntp-services strings              || time server,默认：ntp1.aliyun.com,ntp2.aliyun.com,ntp3.aliyun.com |
|  --taint                             ||   Update the taints on the nodes |

