# 独立ETCD 安装

## 准备主机
|  主机名称  | IP地址 |
| -------- | ------------ |
| etcd1 | 172.16.100.11 |
| etcd2 | 172.16.100.12 |
| etcd3 | 172.16.100.13 |

## 初始化
```shell script
vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13
```
或者使用连续IP方式
```shell script
vik8s etcd init 172.16.100.11-172.16.100.13
```
注意：关于修改节点ssh连接地址，查阅[kubernetes集群初始化](./INSTALL.MD)

## 新节点加入
```shell script
vik8s etcd join 172.16.100.14 172.16.100.15
```