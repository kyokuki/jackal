package modules

import "github.com/ortuman/jackal/component/pubsub/base"

type NodeCreateModule struct {
	
}

func (s *NodeCreateModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleCreate := &base.ElementCriteria{}
	eleCreate.SetName("create")

	eleCrit.AddCriteria(elePubsub).AddCriteria(eleCreate)
	return eleCrit
}
