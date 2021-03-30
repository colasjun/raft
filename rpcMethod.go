package main

import "fmt"

// 投票
func (r *Raft) Vote(n NodeInfo, b *bool) error {
	fmt.Printf("%d投%d开始", r.id, n.ID)
	fmt.Println()

	// bug ??
	//r.setVoteFor(DefaultVoteFor)

	fmt.Println("选举人当前值", r.voteFor, r.leaderId)
	fmt.Println("被选举人", n)

	// 如果已经给别人投过票了 或者 已经有领导了 就不能投了
	if r.voteFor != DefaultVoteFor || r.leaderId != DefaultLeader {
		r.setVoteFor(n.ID)
		fmt.Printf("%d投%d成功", r.id, n.ID)
		fmt.Println()
		*b = true
	} else {
		fmt.Printf("%d投%d失败", r.id, n.ID)
		*b = false
	}

	return nil
}

// 领导者确认
func (r *Raft) ConfirmLeader(n NodeInfo, b *bool) error {
	r.setLeaderId(n.ID)

	*b = true
	fmt.Printf("%d收到%d为领导成功", r.id, n.ID)
	fmt.Println()

	// 恢复默认
	r.setDefault()

	return nil
}

// 心跳检测回复
func (r *Raft) HeartBeatRe (n NodeInfo, b *bool) error {
	r.setLeaderId(n.ID)
	r.lastGetHeartBeatTime = getMillSecond()

	fmt.Printf("%d收到%d领导心跳检测", r.id, n.ID)
	fmt.Println()

	*b = true
	return nil
}

