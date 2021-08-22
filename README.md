# GoR

一个特异化的代理，解决的问题是：需要在家和在外面使用同样一系列域名访问局域网里的机器（比如K8s），但由于电信运营商封锁了443/80端口。

```
局域网内 --> yourdomain.com --> homelab:443

公网 --> yourdomain.com --> gor --> router:custom_port --> homelab:443
```

只要将各种域名解析到gor即可。

还额外加上了刷新ddns IP和接口更新IP的功能。