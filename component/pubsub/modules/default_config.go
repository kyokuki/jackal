package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
)


type DefaultConfigModule struct {}

func (s *DefaultConfigModule)Name() string  {
	return "DefaultConfigModule"
}

func (s *DefaultConfigModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	eleDefault := &base.ElementCriteria{}
	eleDefault.SetName("default")

	elePubsub.AddCriteria(eleDefault)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *DefaultConfigModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#retrieve-default",
	}, nil
}

func (s *DefaultConfigModule)Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError  {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	def := xmpp.NewElementName("default")
	tmpDefaultNodeConfig := base.NewDefaultNodeConfig("default")
	xdefault := tmpDefaultNodeConfig.Form().Element()
	if xdefault == nil {
		errElem := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		errElem.SetAttribute("feature", "config-node")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
			[]xmpp.XElement{
				errElem,
			})
	}

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	def.AppendElement(xdefault)
	pubSubResult.AppendElement(def)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)
	return nil
}
