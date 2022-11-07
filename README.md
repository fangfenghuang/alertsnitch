# 安装步骤
## 安装Mysql（版本5.7，大于5.7有报错，安装过程略），创建数据库alertsnitch，alertsnitch用户及授权
## 创建数据库
create database alertsnitch;
use alertsnitch;

执行建表脚本: db.d/mysql/0.0.1-bootstrap.sql

修改Mode verions数据: db.d/mysql/0.1.0-fingerprint.sql

## 安装alertsnitch，修改下面ALERTSNITCH_DSN的值

ALERTSNITCH_DSN="${MYSQL_USER}:${MYSQL_PASSWORD}(${MYSQL_IP}:${MYSQL_PORT})/{$MYSQL_DATABASE}"

```yaml
# alertsnitch-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertsnitch
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: alertsnitch    
  template:
    metadata:
      labels:
        app.kubernetes.io/name: alertsnitch
    spec:
      containers:
      - image: registry.gitlab.com/yakshaving.art/alertsnitch
        name: alertsnitch
        ports:
        - containerPort: 9567
          name: http
        env:
        - name: ALERTSNITCH_BACKEND
          value: mysql
        - name: ALERTSNITCH_DSN
          value: myuser:123456@tcp(10.19.0.1:3306)/alertsnitch
        readinessProbe:
          httpGet:
            path: /-/ready
            port: 9567
          initialDelaySeconds: 30
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /-/health
            port: 9567
          initialDelaySeconds: 60
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: alertsnitch
  namespace: default
spec:
  ports:
  - name: http
    port: 9567
    targetPort: http
  selector:
    app.kubernetes.io/name: alertsnitch
```

## 配置alertmanager
```yaml
    route:
      receiver: all-to-alertsnitch
      routes:
      #所有告警收集
      - receiver: all-to-alertsnitch
        continue: true
 
    receivers:
    #alertsnitch
    - name: 'all-to-alertsnitch'
      webhook_configs:
      - url: 'http://alertsnitch.default:9567/webhook'
        send_resolved: true
```

# 仪表盘
https://grafana.com/grafana/dashboards/15833

http://alertsnitch.default:9567/-/ready



# 本地调试
go run main.go -dsn "myuser:123456@tcp(10.19.0.1:3306)/alertsnitch" -database-backend null

## webhook
http://<alertsnitch-host-or-ip>:9567/webhook


# 告警历史
## 数据库表
一条告警消息生成一个alertGroupID及AlertID

- alert
一条告警消息记录一条alert

- alertannotation
一条告警消息记录一条alertannotation

- alertgroup
一条告警消息记录一条alertgroup

- alertlabel
一条告警消息记录多条label

- commonanotation
一条告警消息记录一条commonanotation

- commonlabel
一条告警消息记录多条commonlabel

- grouplabel
一条告警消息记录多条grouplabel

- model
alertsnitch版本

## 记录说明
- 未恢复和已恢复告警列表显示相同一组label的alert，如果已恢复则显示结束时间，否则只显示开始时间和持续时间
- 告警历史展示每一次告警消息都会有一个起始时间和结束时间
- 每一次告警消息
- 开始时间是从第一次接收到消息开始
- 如果接收中断，告警会认为是not resolved,但是有记录结束时间
