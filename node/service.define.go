package node

//服务定义
type IServices interface {
	AddEvent(topic string, value interface{})
}
