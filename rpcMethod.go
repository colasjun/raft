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

// 主节点设置值
func (r *Raft) ReceiveMessage(s []string,  b *bool) error {
	// 先获取自己的最大ID
	id := r.msg.getNewId()

	// 记录日志
	r.msg.msgLog[id] = msg{
		act:"set",
		key:s[0],
		val:s[1],
	}

	// 记录应答总数
	ackNum := 0

	// 广播
	go r.broadCast("Raft.SetMessage", s, func(ok bool) {
		if ok {
			ackNum++
		}
	})

	// 判断是否超过半数应答
	for {
		if ackNum > raftNodeCount / 2  - 1{
			r.logger.Println("全网已经超过半数的设置值成功")

			// 自己设置值
			r.msg.set(s[0], s[1])
			// 返回给 客户端
			*b = true
			break
		}
	}

	return nil
}


// 主节点删除值
func (r *Raft) DeleteMessage(s string, b *bool) error {
	// 先获取自己的最大ID
	id := r.msg.getNewId()

	// 记录日志
	r.msg.msgLog[id] = msg{
		act:"delete",
		key:s,
	}

	// 记录应答总数
	ackNum := 0

	// 广播
	go r.broadCast("Raft.DelMessage", s, func(ok bool) {
		if ok {
			ackNum++
		}
	})

	// 判断是否超过半数应答
	for {
		if ackNum > raftNodeCount / 2  - 1{
			r.logger.Println("全网已经超过半数的设置值成功")

			// 自己设置值
			r.msg.delete(s)
			// 返回给 客户端
			*b = true
			break
		}
	}

	return nil
}

// 其他节点设置值
func (r *Raft) SetMessage(s []string, b *bool) error {
	r.msg.set(s[0], s[1])
	*b = true
	return nil
}

// 其他节点删除值
func (r *Raft) DelMessage(key string, b *bool) error {
	r.msg.delete(key)
	*b = true
	return nil
}