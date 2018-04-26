package node

import (
	"github.com/ontio/ontology-eventbus/actor"
)

func NewNeighbors() *Neighbors {
	return &Neighbors{}
}

type Neighbors struct {
	NbrSet   *actor.PIDSet
	retrySet *actor.PIDSet
}

func (n Neighbors) AddNbrNode(v *actor.PID) {
	n.NbrSet.Add(v)
}

func (n Neighbors) DelNbrNode(v *actor.PID) bool {
	return n.NbrSet.Remove(v)
}

func (n Neighbors) GetNbrList() []actor.PID {
	return n.NbrSet.Values()
}

func (n Neighbors) GetNbrNode(name string) (pid *actor.PID) {
	n.NbrSet.ForEach(func(i int, p actor.PID) {
		if pid.GetId() == name {
			pid = &p
			return
		}
	})
	return nil
}

func (n Neighbors) GetNbrAddr(name string) (addr string) {
	n.NbrSet.ForEach(func(i int, pid actor.PID) {
		if pid.GetId() == name {
			addr = pid.GetAddress()
			return
		}
	})
	return
}

func (n Neighbors) AddInRetryList(v *actor.PID) {
}
func (n Neighbors) RemoveFromRetryList(v *actor.PID) {
}

func (n Neighbors) StartNbrWatch() {
}

func (n Neighbors) StopNbrWatch() {
}
