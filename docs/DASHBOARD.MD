# dashboard 安装

简单安装：
```shell script
vik8s sidecars dashboard 
```
安装完成后会打印登录方式。

如果您再次安装完成后，没有记住token您可以通过命令 `vik8s sidecars dashboard` 获取。

## 参数说明

| 参数 | 说明 |
| --- | --- |
| --enable-insecure-login| 是否启用不安全登录方式，启用之后dashboard将不再使用https方式提供服务 |
| --expose | 通过nodePort方式对外提供服务，-1: 禁用, 0: 系统自动分配, >0: 指定端口方式 |
| --ingress | 给dashboard添加入口，默认不添加 |
| --insecure-header | 是否给入口直接添加认证token header这样，进入dashboard就不需要再次输入token了. |
| --tlskey | dashboard.key  证书 |               
| --tlscert | dashboard.crt 证书 |