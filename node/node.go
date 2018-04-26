package node

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/kooksee/uspnet/actors"
	"github.com/ontio/ontology-eventbus/zmqremote"
	"github.com/ontio/ontology-eventbus/mailbox"
	"github.com/kooksee/uspnet/config"
)

func NewNoder(cfg *config.Config) *Noder {
	return &Noder{cfg: cfg}
}

type Noder struct {
	nbr INeighbors
	s   IServices

	cfg *config.Config

	nodePid   *actor.PID
	nodeActor actor.Actor
}

func (n *Noder) Start() error {
	var err error

	zmqremote.Start(n.cfg.GetBindAddr())

	n.nodeActor = &actors.NodeActor{}
	n.nodePid, err = actor.SpawnNamed(actor.FromProducer(func() actor.Actor { return n.nodeActor }).WithMailbox(mailbox.Bounded(1000000)), "node")
	if err != nil {
		return err
	}

	return nil
}

func (n *Noder) InitSeeds() {
}
func (n *Noder) InitNbr() {
	n.nbr = NewNeighbors()
}
func (n *Noder) GetNbr() {
}
func (n *Noder) InitService() {
}
func (n *Noder) GetService() {
}

func (n *Noder) String() {
}

func (n *Noder) Stop() {
	n.nodePid.Stop()
	zmqremote.Shutdonw()
}
