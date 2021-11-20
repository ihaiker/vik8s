![](./docs/logo.png) ![](docs/logo_txt.png)

一个非常简单kubernetes高可用集群安装部署工具，支持 v1.19.+

程序尽可能采用原生kubernetes特性不对kubernetes进行修改和面向过程模式编写，把安装过程清晰化。



## 特性

-   简单快捷方便的安装方式。所有安装基本上就是一条命令
-   多集群管理，方便的管理不同集群。
-   统一命令管理程序，可以方便的在客户端使用一条命令在所有管理主机上运行。
-   独立应用不依赖任何第三方
-   可控的证书时间（默认：44年，本人的幸运数字就是4，我的地盘我任性）
-   可选择性的镜像地址。默认提供国内/外**可信&安全**的镜像地址。不使用离线包和私有镜像（为啥不提供离线包？您是否还记得IOS环境侵入问题，Goolge一下吧，当然这样的话你的所有安装节点必须可以联网去下载镜像。）
-   通过使用service特性和IPVS实现HA高可用，不依赖于任何第三实现。
-   轻松的增加集群节点 `vik8s join -m <ip>`
-   ETCD节点可单独安装和节点添加。`vik8s etcd init <ip1> <ip2> ...` 和 `vik8s etcd join <ip3> ...`
-   提供周边 安装，同样简单方便。
    -   dashboard (1.0 后废弃)
    -   ingress (nginx/traefik)
-   【重磅推出】kubernetes reduce 命令，简化yaml配置文件。 [查看教程和实例](./reduce/index.md)。



## 快速开始：

查看文档：[Quickstart][quickstart]



## 技术支持群

| 钉钉群：34673135         | QQ群：715096758    |
| ------------------------ | ------------------ |
| ![][dd] | ![][qq] |

[dd]: https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=https%3A%2F%2Fqr.dingtalk.com%2Faction%2Fjoingroup%3Fcode%3Dv1%2Ck1%2CZiQs4kjvfFMm5EDwWHPZSGGZRCHeW%2BZUwqlW73xXrO0%3D "钉钉群地址"
[qq]: https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=https%3A%2F%2Fqm.qq.com%2Fcgi-bin%2Fqm%2Fqr%3Fk%3DRif3DnLnNWbZh3BTQod_vYcCpYRSpLkk%26authKey%3DIng0sqF0Wj6o22WMjXBDiu38V9arojuEcy5iGLR%2BfHMNOoYoAkg1tpcK9B3mkL4b%26noverify%3D0 "QQ"
[quickstart]: https://vik8s.renzhen.la/#quick-start
