package node

import "github.com/ontio/ontology-eventbus/actor"

//邻居节点操作
type INeighbors interface {
	// 删除邻居节点
	DelNbrNode(*actor.PID) bool

	// 添加邻居节点
	AddNbrNode(*actor.PID)

	// 获取邻居列表
	GetNbrList() []actor.PID

	// 获得邻居节点
	GetNbrNode(string) *actor.PID

	// 获得节点的地址
	GetNbrAddr(string) string

	// 添加到尝试队列
	//AddInRetryList(name string)

	// 从尝试队列删除
	//RemoveFromRetryList(name string)

	// 启动邻居节点监控
	// 监控邻居节点是否失效
	StartNbrWatch()

	// 停止邻居节点监控
	StopNbrWatch()
}
