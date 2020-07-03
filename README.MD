# vik8s (very easy install kubernetes)
![](./docs/logo.png) ![](./logo.png)

一个非常简单原生多云kubernetes高可用集群安装部署工具，支持 v1.17.+

程序尽可能采用原生kubernetes特性不对kubernetes进行修改和面向过程模式编写，把安装过程清晰化。

[![asciicast](https://asciinema.org/a/342332.svg)](https://asciinema.org/a/342332)

## 特性

- [X] 简单快捷方便的安装方式。所有安装基本上就是一条命令
- [X] 多集群管理，方便的管理不同集群。
- [X] 统一命令管理程序，可以方便的在客户端使用一条命令在所有管理主机上运行。
- [X] 独立应用不依赖任何第三方（尽量吧，毕竟还是需要kubeadm,etcdadm的，尤其是etcdadm还是需要编译安装的，但是这些都是自动处理的）
- [X] 可控的证书时间（默认：44年，本人的幸运数字就是4，我的地盘我任性）
- [X] 可选择性的镜像地址。默认提供国内/外**可信&安全**的镜像地址。不使用离线包和私有镜像（为啥不提供离线包？您是否还记得IOS环境侵入问题，Goolge一下吧，当然这样的话你的所有安装节点必须可以联网去下载镜像。）
- [X] 通过使用service特性和IPVS实现高可用，不依赖于任何第三实现。
- [X] 轻松的增加集群节点 `vik8s join -m <ip>`
- [X] ETCD节点可单独安装和节点添加。`vik8s etcd init <ip1> <ip2> ...` 和 `vik8s etcd join <ip3> ...`
- [X] 提供周边 安装，同样简单方便。
    - [X] dashboard 
    - [X] ingress (nginx/traefik)
    - [ ] 后续会提供更多的安装方式
- [X] <span style="color: red">【重病推出】kubernetes reduce 命令，简化yaml配置文件。 </span> [查看教程和实例](./docs/REDUCE.MD)。
<p style="color:red">Note: 本程序现在仅支持 centos 7/8，是否将来会支持其他系统暂未可知</p>

## 快速开始
> 主机准备

| 主机名称 | IP地址|
|---|---|
| master1 | 172.16.100.10 |
| master2 | 172.16.100.11 |
| master3 | 172.16.100.12 |
| node1 | 172.16.100.13 |
| node2 | 172.16.100.14 |
| node3 | 172.16.100.15 |

>安装

```shell
vik8s init -m 172.16.100.10 -m 172.16.100.11 -m 172.16.100.12 -n 172.16.100.13-172.16.100.15
```

## [安装详细文档](./docs/INSTALL.MD)

## [Reduce配置简化方案](./docs/REDUCE.MD)

## 技术支持群

![](./docs/dd.png) ![](./docs/qq.png) 