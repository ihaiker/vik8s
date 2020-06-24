# Ingress 安装

## Nginx/Treafik Ingress 

nginx:
```shell script
vik8s ingress nginx --hostnetwork --node.selector kubernetes.io/hostname=node1
``` 
traefik:
```shell script
vik8s ingress traefik --hostnetwork --node.selector kubernetes.io/hostname=node1
``` 
执行命令后ingress会使用 hostNetwork方式部署在node1上。其他部署方式可以查阅帮助 `vik8s ingress nginx|traefik --help`
