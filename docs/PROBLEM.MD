# 常见问题汇总
这里归总的一些问题，基本已经在程序安装时处理，但是为了方便还是记录一下。

## 网络问题

### POD无法访问外网或者无法POD之间无法访问。

1、是否是使用了calico这样的网络插件并且开始了网络访问限制。这个就需要自己检查了。通过命令`kubectl get networkpolicies.networking.k8s.io -A`查看是否有网络问题

2、如果 Node 上安装的 Docker 版本大于 1.12，那么 Docker 会把默认的 iptables FORWARD 策略改为 DROP。转发丢弃
  这会引发 Pod 网络访问的问题，在每个节点执行命令 `iptables -P FORWARD ACCEPT`



## Dashboard 使用 `kubectl proxy`无法访问

### 1、提示no endpoints available for service "https:kubernetes-dashboard:"

首先要查看是否是pod未启动成功。`kubectl -n kubernetes-dashboard get pod` 查看状态是不是`Running`状态。
如果不是就需要查看POD为什么没有启动成功，当然了网上的教程大多是images下载错误(气愤的居然google出来的也全是此问题，无奈，😂，😂，😂)。
如果不用的是本软件，此问题当然不存在了。
存在的问题是：本软件对dashboard的yaml文件有所修改增加了一些功能点，尤其是对于service的端口进行了更改，`从443改为8443`和容器统一，所以在官方提供的教程里面就是错误的了需要添加端口号的。
此问题有关的详细内容自己查阅一下kubernates API文档吧。当然了本程序肯定对此进行了处理，这里仅仅是记录一下。

### 2 、如果命令不添加参数需要在安装机器本机本地访问的。

需要添加参数：`--accept-paths='^.*' --address=<hostIP> --accept-hosts="^.*$"`。
但是添加参数以后，在其他主机上通过IP是可以访问了但是登录授权貌似还是有问题的。
这是因为Dashboard只允许localhost和127.0.0.1使用HTTP连接进行访问，而其它地址只允许使用HTTPS。
因此，如果需要在非本机访问Dashboard的话，只能选择其他访问方式：`NodePort`、`Ingress`、公开的`ApiServer`方式。

#### 2.1 NodePort 方式

NodePort方式就是修改server的`type: NodePort`,并且需要你去生成tls证书。对应本程序您只需要添加一个参数 `--expose`。

此参数存在三中类型值

|值|说明|
|---|---|
|-1|参数等于禁用状态，也就是service的type=ClusterIP|
|0|系统自动分配一个端口，并且type=NodePort|
| maxPort > n > 0| 就是NodePort的端口地址 |

#### 2.2 Ingress 方式
网上也有很多教程是针对此种方式的，但是很多教程都没有说其中一个很重要的问题就是证书的信任问题，后端的证书必须是一个。相关介绍请查阅 [TLS communication between Traefik and backend pods¶](1)
有了上面的方式此方式当然也对应本程序的一个参数`--ingress`啦，。

#### 2.3 公开的ApiServer方式。
如果Kubernetes API服务器是公开的，并可以从外部访问，那我们可以直接使用API Server的方式来访问，也是比较推荐的方式。

Dashboard的访问地址为：
```shell script
https://<master-ip>:<apiserver-port>/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/
```
但是返回的结果可能如下：
```json
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {
    
  },
  "status": "Failure",
  "message": "services \"https:kubernetes-dashboard:\" is forbidden: User \"system:anonymous\" cannot get services/proxy in the namespace \"kube-system\"",
  "reason": "Forbidden",
  "details": {
    "name": "https:kubernetes-dashboard:",
    "kind": "services"
  },
  "code": 403
}
```

这是因为最新版的k8s默认启用了RBAC，并为未认证用户赋予了一个默认的身份：anonymous。对于API Server来说，它是使用证书进行认证的，我们需要先创建一个证书：

- 1.首先找到kubectl命令的配置文件，默认情况下为/etc/kubernetes/admin.conf，
- 2.然后我们使用client-certificate-data和client-key-data生成一个p12文件，可使用下列命令：

```shell
#生成client-certificate-data
grep 'client-certificate-data' ~/.kube/config | head -n 1 | awk '{print $2}' | base64 -d >> kubecfg.crt

# 生成client-key-data
grep 'client-key-data' ~/.kube/config | head -n 1 | awk '{print $2}' | base64 -d >> kubecfg.key

# 生成p12，
openssl pkcs12 -export -clcerts -inkey kubecfg.key -in kubecfg.crt -out kubecfg.p12 -name "kubernetes-client"

# 导入P12文件就可以打开了
```

### 3、dashboard进去后显示 `square/go-jose: error in cryptographic primitive`
这个就是简单问题了。你只需要点击右上角的注销重新加来就好了。因为dashboard把你之间的token记住了，而你又重新部署过token更换了


## 节点加入后 `Warning  ImageGCFailed  kubelet, xxx failed to get imageFs info: unable to find data in memory cache`
处理方式 
```shell script
yum install -y systemd
```
这样一个一个的安装比较费事，我们的工具就来了，在全部节点运行相同命令
```shell script
vik8s -- yum install -y systemd
```

## GlusterFS 

### 集群 复制分数问题
集群 brick % 复制分数 == 0

### 新接待加入时使用hostname问题
在新加入节点使用hostname方式会出现一些问题，提示无法连接peer，
这是因为在gluster完成第一次安装后后面又有新机器添加，如果使用hostname方式的话需要添加/etc/hosts文件内容，之前初始化的pod里面并不包含这些新添加的机器
解决方式：

- 更新pod，同时会自动更新/etc/hosts文件(会导致peer断联几秒钟)
- 直接更新host文件. `cat /etc/hosts >> /var/lib/kubelet/pods/$(PODID)/etc-hosts`

## 最后记录一个golang方法特别容易出的问题（BUG 我觉得应该算是吧）,因为这个导致我错误好几次呢。
问题描述：在一个yaml文件中，总共有 44 行。执行下面的方法结果是多少? 感兴趣的朋友可以试验一下。我可以告诉你的是结果不一定是 44，至于是为什么看一下源码就了解了

```golang 

func FileLine(filePath string ) {
	f , err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	num := 0
	for {
		_, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		num += 1
	}
	fmt.Println(num)
}

```

[1]: https://docs.traefik.io/v1.7/configuration/backends/kubernetes/#tls-communication-between-traefik-and-backend-pods "TLS communication between Traefik and backend pods¶"