# raft
raft算法实现

整个raft节点写死的 没有实现动态扩展 核心只实现raft算法 + 数据存储 主要是为了加深理解raft

###
第一步构建 
````
go build -o raft ./*.go
````

第二步分别启动三台服务 
```
./raft 1 
./raft 2
./raft 3

```

第三步访问服务设置值
```
http://localhost:8011/set?key=test&val=testValue
http://localhost:8012/get?key=test
http://localhost:8013/delete?key=test
```
