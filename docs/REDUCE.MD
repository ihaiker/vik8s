# Reduce kubernetes 简化配置

无论是在学习或者工作中都要编写yaml文件，我不知道其他人对yaml配置是否存在这样困惑：配置是否有些复杂了（当然了解了细节也明白是为啥）。
本程序reduce命令将尽可能的简化配置。

reduce命令采用类似`nginx.conf`的配置方式配置kubernetes。

先给一个简单的配置：mysql.conf
```nginx
# 配置kubernetes版本
kubernetes v1.18.2;

# 配置自动添加label的前缀，默认：vik8s.io
prefix vik8s.io;

namespace vik8s; 

configmap mysql-root-password {
    mysql.root.password haiker;
}

configmap mysql-vik8s-password {
    mysql.vik8s.password vik8s;
}



```

执行编译后的文件：`vik8s reduce mysql.conf -o mysql.yaml`
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  namespace: ext
  labels:
    app: mysql
    tier: backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
      tier: backend
  template:
    metadata:
      name: mysql
      labels:
        app: mysql
        tier: backend
    spec:
      restartPolicy: Always
      containers:
        - name: mysql
          image: mysql:5.7.29
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3306
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                configMapKeyRef:
                  name: mysql
                  key: mysql.root.password
            - name: MYSQL_DATABASE
              value: extez
            - name: MYSQL_USER
              value: ext
            - name: MYSQL_PASSWORD
              valueFrom:
                configMapKeyRef:
                  name: mysql
                  key: mysql.ext.password
          envFrom:
            - configMapRef:
                name: mysql
          volumeMounts:
            - mountPath: "/var/lib/mysql"
              name: mysql-data
      volumes:
        - name: mysql-data
          persistentVolumeClaim:
            claimName: gluster-mysql-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: ext
spec:
  selector:
    app: mysql
    tier: backend
  ports:
    - port: 3306
      targetPort: 3306
  type: ClusterIP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-admin
  namespace: ext
  labels:
    app: mysql-admin
    tier: frontend
spec:
  selector:
    matchLabels:
      app: mysql-admin
      tier: frontend
  template:
    metadata:
      name: mysql-admin
      labels:
        app: mysql-admin
        tier: frontend
    spec:
      containers:
        - name: phpmyadmin
          image: phpmyadmin/phpmyadmin
          imagePullPolicy: IfNotPresent
          env:
            - name: PMA_HOST
              value: mysql
            - name: PMA_USER
              value: root
            - name: PMA_PASSWORD
              valueFrom:
                configMapKeyRef:
                  key: mysql.root.password
                  name: mysql
          ports:
            - name: http
              containerPort: 80
              hostPort: 80
          envFrom:
            - configMapRef:
                name: mysql

---
apiVersion: v1
kind: Service
metadata:
  name: mysql-admin
  namespace: ext
spec:
  type: ClusterIP
  selector:
    app: mysql-admin
    tier: frontend
  ports:
    - name: http
      port: 80
      targetPort: 80

---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: mysql-admin
  namespace: ext
spec:
  rules:
    - host: db.extez.com
      http:
        paths:
          - backend:
              serviceName: mysql-admin
              servicePort: http

```

