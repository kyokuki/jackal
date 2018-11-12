package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"strings"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type PublishItemModule struct{}

func (s *PublishItemModule) Name() string {
	return "PublishItemModule"
}

func (s *PublishItemModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	elepublish := &base.ElementCriteria{}
	elepublish.SetName("publish")

	elePubsub.AddCriteria(elepublish)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *PublishItemModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#publish",
	}, nil
}

func (s *PublishItemModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	publish := pubsub.Elements().Child("publish")
	nodeName := strings.Trim(publish.Attributes().Get("node"), " ")
	id := packet.ID()
	//stanzaType := strings.ToLower(packet.Attributes().Get("type"))

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	publishResult := xmpp.NewElementName("publish")

	if nodeName == "" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("node-required", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	} else {
		if nodeConfig.GetNodeType() == enums.Collection {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
				[]xmpp.XElement{
					xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors"),
				})
		}
	}


	pubSubResult.AppendElement(publishResult)
	resultStanza.AppendElement(pubSubResult)
	//stm.SendElement(resultStanza)
	return nil
}
