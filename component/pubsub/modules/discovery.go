package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"fmt"
	"github.com/ortuman/jackal/module"
)

type DiscoveryModule struct{}

func (s *DiscoveryModule) Name() string {
	return "DiscoveryModule"
}

func (s *DiscoveryModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")

	eleQuery := &base.OrElementCriteria{}
	eleQuery1 := &base.ElementCriteria{}
	eleQuery1.SetName("query").AddAttr("xmlns", "http://jabber.org/protocol/disco#info")
	eleQuery2 := &base.ElementCriteria{}
	eleQuery2.SetName("query").AddAttr("xmlns", "http://jabber.org/protocol/disco#items")
	eleQuery.AddCriteria(eleQuery1)
	eleQuery.AddCriteria(eleQuery2)

	eleCrit.AddCriteria(eleQuery)
	return eleCrit
}

func (s *DiscoveryModule) Process(stanza xmpp.XElement, stm stream.C2S) {
	fmt.Println(s.Name() + " process ")

	query := stanza.Elements().Child("query")
	node := query.Attributes().Get("node")
	xmlns := query.Namespace()

	if node == "" {
		fmt.Println(xmlns)
		switch stanza := stanza.(type) {
		case *xmpp.IQ:
			if di := module.Modules().DiscoInfo; di != nil && di.MatchesIQ(stanza) {
				di.ProcessIQ(stanza, stm)
				return
			}
		}
		return
	}

	if xmlns == "http://jabber.org/protocol/disco#info" {
		s.processDiscoInfo(stanza, stm)
	}

	if xmlns == "http://jabber.org/protocol/disco#items" {
		s.processDiscoItems(stanza, stm)
	}
}

func (s *DiscoveryModule) processDiscoInfo(stanza xmpp.XElement, stm stream.C2S) {
	stan := stanza.(xmpp.Stanza)
	fromJID := stan.FromJID()
	toJID := stan.ToJID()

	resultIq := xmpp.NewElementName(stanza.Name())
	resultIq.SetTo(fromJID.String())
	resultIq.SetFrom(toJID.String())
	resultQuery := xmpp.NewElementNamespace("query", "http://jabber.org/protocol/disco#info")

	resultIq.AppendElement(resultQuery)
	stm.SendElement(resultIq)
}

func (s *DiscoveryModule) processDiscoItems(stanza xmpp.XElement, stm stream.C2S) {
	stan := stanza.(xmpp.Stanza)
	fromJID := stan.FromJID()
	toJID := stan.ToJID()

	resultIq := xmpp.NewElementName(stanza.Name())
	resultIq.SetTo(fromJID.String())
	resultIq.SetFrom(toJID.String())
	resultQuery := xmpp.NewElementNamespace("query", "http://jabber.org/protocol/disco#items")

	resultIq.AppendElement(resultQuery)
	stm.SendElement(resultIq)
}