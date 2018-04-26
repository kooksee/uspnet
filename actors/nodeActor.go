package actors

import (
	"github.com/ontio/ontology-eventbus/actor"
	"fmt"
)

type NodeActor struct {
}

func (n *NodeActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize server actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *Ping:

	}
}

type Ping struct {
}
