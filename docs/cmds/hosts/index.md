# 主机访问方式管理



在某些情况下，会出现每个主机的密码或秘钥不一致，或者集群访问需要通过代理，为了让安装更方便，我们添加了 `hosts`命令。此命令可以先行把集群ip添加的到管理之中，在之后指定主机的时候，我们只需要指定主机的`IP`或者`hostname`即可。

## 简单实例

```shell
$ vik8s hosts -u root -P passwd 10.24.0.10
$ vik8s hosts -u admin -i ~/.ssh/id_rsa 10.24.0.11
$ vik8s hosts -u addmin -P passwd --port 22
```

命令基本上符合`ssh`命令的参数，就不多余赘述。有关参数可以使用 `vik8s hosts --help`方式查看。



## 参数详解

| 参数              | 默认值            | 环境变量获取          | 描述                                |
| ----------------- | ----------------- | --------------------- | ----------------------------------- |
| -u, –user         | root              | VIK8S_SSH_USER        | 设置登录账户                        |
| -P, --password    |                   | VIK8S_SSH_PASSWORD    | ssh账户密码                         |
| --port            | 22                | VIK8S_SSH_PORT        | ssh端口号                           |
| -i, --private-key | $HOME/.ssh/id_rsa | VIK8S_SSH_PRIVATE_KEY | ssh private key                     |
| --proxy           |                   | VIK8S_SSH_PROXY       | 使用的ssh代理，更多查看代理使用章节 |
| --passphras       |                   | VIK8S_SSH_PASSPHRASE  | private key passphrase              |



## 使用代理

为了保证系统安全有些时候我们无法直接访问控制节点，需要使用代理去访问此时我们就需要使用需要使用ssh代理访问我们的主机。

添加代理和添加主机一样，我们主要先把代理ssh主机10.24.加入到管理列表，

```shell
vik8s hosts -u root -P password <ssh_proxy_ip>
```

然后添加我们需要访问的主机

```shell
vik8s hosts <node_ip> --proxy=<ssh_proxy_ip>
```



## 使用技巧

为了让管理变得更通用，我们进一步支持了`ssh`命令的一致性

```shell
vik8s hosts admin@10.24.0.10:12232 
```



