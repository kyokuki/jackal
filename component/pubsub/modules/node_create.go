package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
)


type NodeCreateModule struct {}

func (s *NodeCreateModule)Name() string  {
	return "NodeCreateModule"
}

func (s *NodeCreateModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleCreate := &base.ElementCriteria{}
	eleCreate.SetName("create")

	elePubsub.AddCriteria(eleCreate)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *NodeCreateModule)Process(stanza xmpp.XElement, stm stream.C2S)  {
	
}
