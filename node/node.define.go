package node

type INoder interface {
	// 节点启动
	Start() error

	// 节点停止
	Stop() error

	// 初始化seeds
	InitSeeds() error

	// 初始化邻居节点
	InitNbr() error

	// 获的邻居节点
	GetNbr()

	// 初始化服务
	InitService() error

	// 获的服务
	GetService()

	// 序列化
	String()
}
