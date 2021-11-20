# Reduce kubernetes 简化配置

无论是在学习或者工作中都要编写yaml文件，我不知道其他人对yaml配置是否存在这样困惑：配置是否有些复杂了（当然了解了细节也明白是为啥）。
本程序reduce命令将尽可能的简化配置。

reduce命令采用类似`nginx.conf`的配置方式配置kubernetes。

先给一个简单的配置：mysql.conf
```nginx
kubernetes v1.18.2;
prefix vik8s.io;
namespace vik8s;

configmap mysql-config {
    mysql.root.password haiker;
    mysql.vik8s.password vik8s;
}

deployment mysql {
    restartPolicy Always;
    container mysql mysql:5.7.29 IfNotPresent {
        port 3306;
        env MYSQL_ROOT_PASSWORD configmap mysql-config mysql.root.password;
        envs {
            MYSQL_DATABASE vik8s;
            MYSQL_USER vik8s;
            MYSQL_PASSWORD configmap mysql-config mysql.vik8s.password;
        }
        mount hostPath:mysql-data /data/mysql:/var/lib/mysql;
    }

    container php-my-admin phpmyadmin/phpmyadmin {
        envs {
            PMA_HOST 127.0.0.1;
            PMA_USER root;
            PMA_PASSWORD configmap mysql-config mysql.root.password;
        }
        port http 80;
    }
}

service deployment:mysql mysql-admin {
    port admin 80:80;
    port mysql 3306:3306;
}

ingress mysql-admin {
    rules myadmin.vik8s.io {
        http paths {
            serviceName mysql-admin;
            servicePort admin;
        }
    }
}
```
使用 `vik8s reduce mysql.conf` 编译后输出yaml内容。不知道您是否觉得会感觉配置清爽了许多呢（只要您不回答：呵呵，就好）。

## Reduce 原理

配置分为四种配置方式：

  - [用户自定配置插件](./plugins.md)
  - Reduce 系统定义配置方式
  - kubernetes配置转换方式
  - yaml标签方式。

优先顺序为从前向后。

### Reduce 系统定义配置方式

关于此处就不使用长篇论述了，您可以自己使用 `vik8s reduce demo <kind>` 查看demo实例。    
    
### kubernetes配置转换方式
由于kubernetes kind比较多，所有并非所有配置都做了简化处理，当然如果依然要使用此方式配置的话，就需要提供一个配置方式。此方式应用而生。
此方式配置有个统一的前缀 `kind:version name`。下面我们就一个简单的实例来说明。

```nginx 
CustomResourceDefinition:apiextensions.k8s.io/v1beta1 bgpconfigurations.crd.projectcalico.org {
    scope Cluster;
    group crd.projectcalico.org;
    version v1;
    names {
        kind BGPConfiguration;
        plural bgpconfigurations;
        singular bgpconfiguration;
    }
}


PodSecurityPolicy:policy/v1beta1 psp.flannel.unprivileged {
    annotations {
        seccomp.security.alpha.kubernetes.io/allowedProfileNames docker/default;
        seccomp.security.alpha.kubernetes.io/defaultProfileName docker/default;
        apparmor.security.beta.kubernetes.io/allowedProfileNames runtime/default;
        apparmor.security.beta.kubernetes.io/defaultProfileName runtime/default;
    }
    privileged false;
    volumes configMap secret emptyDir hostPath;

    allowedHostPaths {
        pathPrefix "/etc/cni/net.d";
    }
    allowedHostPaths pathPrefix=/etc/kube-flannel;
    allowedHostPaths pathPrefix=/run/flannel;
    
    readOnlyRootFilesystem false;
    runAsUser {
        rule RunAsAny;
    }
    supplementalGroups {
        rule RunAsAny;
    }
    fsGroup rule=RunAsAny;
    # Privilege Escalation
    allowPrivilegeEscalation false;
    defaultAllowPrivilegeEscalation false;
    # Capabilities
    allowedCapabilities 'NET_ADMIN';
    # Host namespaces
    hostPID false;
    hostIPC false;
    hostNetwork true;
    hostPorts {
        min 0;
        max 65535;
    }
    # SELinux
    seLinux rule=RunAsAny;
}
```

### yaml标签方式。
此方式就是直接使用yaml配置文件，做配置。
```nginx
yaml '
yamlSource
';
```

## 更多Demo查看
```shell 
vik8s reduce demo
```
