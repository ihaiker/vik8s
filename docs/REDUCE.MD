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
- [用户自定配置插件](./REDUCE.PLUGINS.MD)
- Reduce 系统定义配置方式
- kubernetes配置转换方式
- yaml标签方式。

优先顺序为从前向后。

### Reduce 系统定义配置方式

### kubernetes配置转换方式

### yaml标签方式。

## 更多Demo