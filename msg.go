package main

import "sync"


type msg struct {
	act string
	key string
	val string
}

type msgStore struct {
	maxId int // 当前最大ID
	msgLog map[int]msg //维护的消息日志
	msg map[string]string //实际消息
	lock sync.Mutex //锁
}

func (m *msgStore) getNewId() int {
	m.lock.Lock()
	m.maxId = m.maxId + 1
	m.lock.Unlock()
	return m.maxId
}

func (m *msgStore) get(key string) string  {
	if val, exist := m.msg[key] ; exist {
		return val
	}
	return "null"
}

func (m *msgStore) set(key string, msg string) bool  {
	m.lock.Lock()
	m.msg[key] = msg
	m.lock.Unlock()
	return true
}

func (m *msgStore) delete(key string) bool  {
	m.lock.Lock()
	delete(m.msg, key)
	m.lock.Unlock()
	return true
}