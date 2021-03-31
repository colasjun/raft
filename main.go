package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// 选举超时时间
var voteTimeOut = 3 // 5s

// 节点总数量
var raftNodeCount = 3

// 心跳检测频率
var heartTimes = 3

// 心跳检测超时时间
var heartTimeOut = 7

// 对外提供服务检测时间
var supportTimes = 2

// node结构表(节点之间通讯）
var nodeTable map[int]string

//节点池（对外提供服务）
var nodeTableHttp map[int]string

// rpc协议
var rpcProtocol = "tcp"

func main()  {

	// 这里为了方便 先写死3个node 后续提供统一注册中心?? （实现节点之间通讯的端口）
	nodeTable = map[int]string {
		1 : "9111",
		2 : "9112",
		3 : "9113",
	}

	//定义三个节点  节点编号 - 监听端口号（对外提供服务端口）
	nodeTableHttp = map[int]string{
		1: "8011",
		2: "8012",
		3: "8013",
	}


	// 判断当前连接的机器是否属于集群
	if len(os.Args) != 2 {
		panic("参数不正确")
	}

	i, _ := strconv.Atoi(os.Args[1])

	if _,ok := nodeTable[i] ; !ok {
		panic("不存在这个服务器")
	}

	if _,ok := nodeTableHttp[i] ; !ok {
		panic("不存在这个服务器")
	}

	// 创建raft
	raft := NewRaft("127.0.0.1", nodeTable[i], i, nodeTableHttp[i])
	//nodeTable[i] = raft

	// 注册rpc
	go rpcRegister(raft)

	// 开启心跳检测
	go raft.heartBeat()

	// 对外提供服务
	go raft.httpListen()

	// 开启选举
	CirCle:
		go func() {
			for {
				// 推荐自己成为候选人
				if raft.becomeCandidate() {
					// 发起选举
					if raft.election() {
						break
					} else {
						fmt.Printf("节点%d选举继续", raft.id)
						fmt.Println()
						//time.Sleep(time.Second * time.Duration(RandInt64(1,10)))
						continue
					}
				} else {
					//time.Sleep(time.Second * time.Duration(RandInt64(1,10)))
					break
				}
			}
		}()

		for {
			// 每秒检测一次
			time.Sleep(time.Second * time.Duration(2))
			//fmt.Println("心跳检测是否超时", raft.id, raft.lastGetHeartBeatTime, getMillSecond(), int64(raft.timeout*1000))
			if raft.lastGetHeartBeatTime != 0 && getMillSecond() - raft.lastGetHeartBeatTime > int64(raft.timeout*1000) {
				fmt.Printf("%d心跳超时,重新选举", raft.id)
				fmt.Println()
				raft.setDefault()
				raft.setLeaderId(DefaultLeader)
				raft.lastGetHeartBeatTime = 0
				goto CirCle
			}
		}
}

