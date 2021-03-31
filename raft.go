package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	DefaultLeader int = -1
	DefaultVoteFor int = -1
	DefaultVoteNum int = 0
	DefaultCurrentTerm = 0

	FollowerStatus int = 0
	CandidateStatus int = 1
	LeaderStatus = 2
)

// 定义raft中每个节点的结构
type Raft struct {
	// ID
	id int

	// 节点信息
	node *NodeInfo

	// 节点状态 身份  0 follower  1 candidate  2 leader
	status int

	// 节点获得的投票数量
	getVoteNum int

	// 当前节点是否已经投过票
	voteFor int

	// 当前节点任期
	currentTerm int

	// 锁
	lock sync.Mutex

	// 当前节点的领导
	leaderId int

	// 心跳超时时间
	timeout int

	// 心跳通道
	heartCh chan bool

	// 接收投票通道
	voteCh chan bool

	// 收到最后一天心跳消息的时间
	lastGetHeartBeatTime int64

	// 自己维护的消息体
	msg *msgStore

	// 日志logger
	logger *log.Logger
}

type NodeInfo struct {
	ID int // 身份ID
	IP string // IP
	PORT string // 端口
	HttpPort string //对外提供服务的端口
}

// 实例化raft节点
func NewRaft(IP string, PORT string, ID int, HttpPORT string) *Raft{
	// 先实例化节点信息
	node := &NodeInfo{
		IP:IP,
		PORT:PORT,
		ID:ID,
		HttpPort:HttpPORT,
	}

	// 实例化Raft对象
	raft := new(Raft)

	// id
	raft.id = node.ID

	// 节点信息
	raft.node = node

	// 节点获取到的投票数
	raft.setGetVoteNum(DefaultVoteNum)

	// 没有给人通过票
	raft.setVoteFor(DefaultVoteFor)

	// 节点任期
	raft.setCurrentTerm(DefaultCurrentTerm)

	// 当前节点状态 follower
	raft.setStatus(FollowerStatus)

	// 暂无leader
	raft.setLeaderId(DefaultLeader)

	// 投票通道
	raft.voteCh = make(chan bool)

	// 心跳通道
	raft.heartCh = make(chan bool)

	// 心跳检测超时时间
	raft.timeout = heartTimeOut

	// 日志logger
	raft.logger = createLogger()

	// 先实例化自己的消息
	msgs := &msgStore{
		maxId:0,
		msgLog:make(map[int]msg, 20),
		msg:make(map[string]string, 20),
	}
	//msgs.msg["test"] = "我说测试数据"

	raft.msg = msgs
	return raft

}

// 修改raft节点为候选人
func (r *Raft) becomeCandidate() bool  {
	// 睡眠下在成为候选人吧
	rand := RandInt64(1500, 5000)
	fmt.Printf("节点%d准备候选人,sleep %d ms", r.id, rand)
	fmt.Println()
	time.Sleep(time.Duration(rand) * time.Millisecond)

	// 判断当前节点已经投过票 或者已经存在领导 不用在成为候选人了
	if r.status == 0 && r.leaderId == DefaultLeader && r.voteFor == DefaultVoteFor {
		// 状态修改为 候选人
		r.setStatus(CandidateStatus)

		// 为自己投票
		r.setVoteFor(r.id)

		// 任期+1
		r.setCurrentTerm(r.currentTerm + 1)

		// 当前没有leader
		r.setLeaderId(DefaultLeader)

		// 自己投票数+1
		r.addGetVoteNum()

		fmt.Printf("节点%d已经变成候选人,得票数量%d", r.id, r.getVoteNum)
		fmt.Println()

		return true
	} else {
		// 成为候选人失败
		fmt.Printf("节点%d成为候选人失败", r.id)
		fmt.Println()
		return false
	}
}

// 进行选举
func (r *Raft) election() bool  {
	fmt.Printf("节点%d开始进行选举,并向其他节点广播", r.id)
	fmt.Println()

	// 广播
	go r.broadCast("Raft.Vote", r.node, func(ok bool) {
		fmt.Printf("Raft.Vote to result %v", ok)
		fmt.Println()
		r.voteCh <- ok
	})

	// 开始循环
	for {
		select {
		case <-time.After(time.Second * time.Duration(voteTimeOut)):
			fmt.Printf("节点%d选举超时", r.id)
			fmt.Println()
			r.setDefault()
			fmt.Printf("节点%d选举超时,设置选举的默认值为%d", r.id, r.voteFor)
			fmt.Println()
			return false
		case ok := <- r.voteCh:
			fmt.Printf("节点%d收到选举结果%v", r.id, ok)
			fmt.Println()
			if ok {
				// 投票+1
				r.addGetVoteNum()
				fmt.Printf("节点%d获得来自其他节点的投票，当前得票数：%d\n", r.id, r.getVoteNum)
			}

			// 投票超过一半并且没有领导 那你就来当领导吧
			if r.getVoteNum > raftNodeCount / 2 && r.leaderId == DefaultLeader {
				fmt.Printf("节点%d得到一半以上的(%d)的投票,选举为领导", r.id, r.getVoteNum)
				fmt.Println()

				// 状态改为领导
				r.setStatus(LeaderStatus)

				// 领导是自己
				r.setLeaderId(r.id)

				// 向其他节点广播
				go r.broadCast("Raft.ConfirmLeader", r.node, func(ok bool) {
					fmt.Println("Raft.ConfirmLeader result", ok)
				})

				// 开启心跳检测？？
				r.heartCh <- true

				return true
			} else {
				fmt.Printf("节点%d没有得到一半以上的(%d)的投票", r.id, r.getVoteNum)
				fmt.Println()
			}

		}
	}
}

// 心跳方法
func (r *Raft) heartBeat() {
	// 如果开启了通道 想其他所有节点开始心跳检测
	if <- r.heartCh {
		for {
			fmt.Printf("领导节点(%d)开始向其他节点发起心跳检测", r.id)
			fmt.Println()
			r.broadCast("Raft.HeartBeatRe", r.node, func(ok bool) {
				fmt.Println("收到回复:", ok)
			})

			// 最后发送消息的时间
			r.lastGetHeartBeatTime = getMillSecond()

			// 心跳检测休眠
			time.Sleep(time.Second * time.Duration(heartTimes))
		}
	}
}

/////////////////////////////保证并发安全的相关方法
func (r *Raft) setStatus (status int)  {
	r.lock.Lock()
	r.status = status
	r.lock.Unlock()
}

func (r *Raft) setLeaderId (leaderId int)  {
	r.lock.Lock()
	r.leaderId = leaderId
	r.lock.Unlock()
}

func (r *Raft) setCurrentTerm (currentTerm int)  {
	r.lock.Lock()
	r.currentTerm = currentTerm
	r.lock.Unlock()
}

func (r *Raft) setVoteFor (voteFor int)  {
	r.lock.Lock()
	r.voteFor = voteFor
	r.lock.Unlock()
}

func (r *Raft) setGetVoteNum (voteNum int)  {
	r.lock.Lock()
	r.getVoteNum = voteNum
	r.lock.Unlock()
}

func (r *Raft) addGetVoteNum ()  {
	r.lock.Lock()
	r.getVoteNum++
	r.lock.Unlock()
}

// 有了领导之后恢复默认值
func (r *Raft) setDefault() {
	//r.lock.Lock()

	// 没有得票
	r.setGetVoteNum(DefaultVoteNum)

	// 不给任何人投票
	r.setVoteFor(DefaultVoteFor)

	// 状态为追随者
	r.setStatus(FollowerStatus)

	//r.lock.Unlock()
}




