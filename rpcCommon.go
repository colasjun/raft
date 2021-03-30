package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)


// rpc服务注册
func rpcRegister (raft *Raft)  {
	/*fmt.Println("rpc服务注册开始")
	err := rpc.Register(r)

	if err != nil {
		fmt.Printf("%d[%s]rpc服务注册失败" + err.Error(), r.id, r.node.IP + ":" + r.node.PORT)
		return
	}
	//把服务绑定到http协议上
	rpc.HandleHTTP()
	err = http.ListenAndServe(":" + r.node.PORT, nil)
	if err != nil {
		fmt.Printf("%d[%s]rpc服务注册失败" + err.Error(), r.id, r.node.IP + ":" + r.node.PORT)
		return
	}*/
	//注册一个服务器
	fmt.Println("rpc服务注册开始")
	err := rpc.Register(raft)
	if err != nil {
		log.Panic(err.Error())
	}
	port := raft.node.PORT
	//把服务绑定到http协议上
	rpc.HandleHTTP()
	//监听端口
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Panic("注册rpc服务失败" + err.Error())
		//fmt.Println("注册rpc服务失败", err)
	}
	fmt.Println("注册rpc服务成功", raft.node)


	/*rpc.RegisterName("Raft", r)

	listener, err := net.Listen("tcp", ":" + r.node.PORT)
	if err != nil {
		fmt.Printf("%d[%s]rpc服务注册失败" + err.Error(), r.id, r.node.IP + ":" + r.node.PORT)
		return
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("%d[%s]rpc服务注册失败" + err.Error(), r.id, r.node.IP + ":" + r.node.PORT)
		return
	}

	rpc.ServeConn(conn)
	fmt.Printf("%d[%s]rpc服务注册成功", r.id, r.node.IP + ":" + r.node.PORT)
	fmt.Println()*/
}


// 广播方法
func (r *Raft) broadCast (method string, args interface{}, fun func(ok bool)) {
	// 不要给自己广播就行

	//fmt.Println("广播所有信息" , nodeTable)

	for id,port := range nodeTable {
		//fmt.Println("广播ID", id)
		if id == r.id {
			//fmt.Println("广播ID是自己,跳过继续", id)
			continue
		}

		// 远程rpc调用
		client, err := rpc.DialHTTP(rpcProtocol, "127.0.0.1:" + port)
		if err != nil {
			fmt.Printf("节点%d向节点%d广播连接失败" + err.Error(), r.id, id)
			fmt.Println()
			fun(false)
			continue
		}
		defer client.Close()

		var bo = false
		err = client.Call(method, args, &bo)
		if err != nil {
			fmt.Printf("节点%d向节点%d广播调用方法失败", r.id, id)
			fun(false)
			continue
		}

		fmt.Printf("节点%d向节点%d广播成功", r.id, id)
		fmt.Println()
		fun(bo)
	}
}


