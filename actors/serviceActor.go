package actors

import (
	"github.com/ontio/ontology-eventbus/actor"
	"fmt"
)

type ServiceActor struct {
}

func (s *ServiceActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize server actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	}
}

func (s *ServiceActor) Start() *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return s })
	pid := actor.Spawn(props)
	return pid
}

func (s *ServiceActor) Stop(pid *actor.PID) {
	pid.Stop()
}
