package main

import (
	"net/http"
	"net/rpc"
	"time"
)

// 每个raft节点都应该实现对外服务
func (r *Raft) httpListen () {

	// 为了保证选举ok sleep3
	time.Sleep(time.Second * time.Duration(3))

	r.logger.Printf("%d开启对外服务,端口%s", r.id, r.node.HttpPort)

	// 因为这里不同的http服务 在本地部署 所以端口只能不一样 实际上分布式节点在不同的服务上面 端口应该一致

	// 查询
	http.HandleFunc("/get" , r.get)

	// 设置 add + modify
	http.HandleFunc("/set" , r.set)

	// 删除
	http.HandleFunc("/delete" , r.delete)

	if serve := http.ListenAndServe(":" + r.node.HttpPort, nil) ; serve != nil {
		r.logger.Printf("%d开启对外服务,端口%s 失败", r.id, r.node.HttpPort)
		r.logger.Println()
	}
}

// 查询服务任何机器都可以直接查询
func (r *Raft) get(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.Form["key"]) < 1 {
		writer.Write([]byte("null"))
		return
	}

	key := request.Form["key"][0]
	r.logger.Printf("%d 收到get key = %s", r.id, key)
	// 直接返回数据即可
	writer.Write([]byte(r.msg.get(key)))
	return
}


// 设置值
func (r *Raft) set(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	if len(request.Form["key"]) < 1 || len(request.Form["val"]) < 1{
		writer.Write([]byte("params error"))
		return
	}

	if r.leaderId == DefaultLeader {
		writer.Write([]byte("inner error"))
		return
	}

	key := request.Form["key"][0]
	val := request.Form["val"][0]
	r.logger.Printf("%d 收到set %s = %s", r.id, key, val)


	// 开始设置值 全部调用leaderRpc
	// 远程rpc调用
	client, err := rpc.DialHTTP(rpcProtocol, "127.0.0.1:" + nodeTable[r.leaderId])
	if err != nil {
		writer.Write([]byte("dial leader error"))
		return
	}
	defer client.Close()

	re := "inner error"
	bo := false
	err = client.Call("Raft.ReceiveMessage", []string{key,val}, &bo)
	if err != nil {
		r.logger.Printf("节点%d通知节点%d失败:%s", r.id , r.leaderId,  err.Error())
		writer.Write([]byte(re))
	}

	if bo {
		writer.Write([]byte("true"))
	}

	return
}



func (r *Raft) delete(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	if len(request.Form["key"]) < 1{
		writer.Write([]byte("params error"))
		return
	}

	if r.leaderId == DefaultLeader {
		writer.Write([]byte("inner error"))
		return
	}

	key := request.Form["key"][0]
	r.logger.Printf("%d 收到delete %s", r.id, key)


	// 开始设置值 全部调用leaderRpc
	// 远程rpc调用
	client, err := rpc.DialHTTP(rpcProtocol, "127.0.0.1:" + nodeTable[r.leaderId])
	if err != nil {
		writer.Write([]byte("dial leader error"))
		return
	}
	defer client.Close()

	re := "inner error"
	bo := false
	err = client.Call("Raft.DeleteMessage", key , &bo)
	if err != nil {
		r.logger.Printf("节点%d通知节点%d失败:%s", r.id , r.leaderId,  err.Error())
		writer.Write([]byte(re))
	}

	if bo {
		writer.Write([]byte("true"))
	}

	return
}