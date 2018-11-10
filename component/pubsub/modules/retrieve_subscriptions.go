package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"strings"
	"log"
)

type RetrieveSubscriptionsModule struct{}

func (s *RetrieveSubscriptionsModule) Name() string {
	return "RetrieveSubscriptionsModule"
}

func (s *RetrieveSubscriptionsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleSubscriptions := &base.ElementCriteria{}
	eleSubscriptions.SetName("subscriptions")

	elePubsub.AddCriteria(eleSubscriptions)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *RetrieveSubscriptionsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#retrieve-subscriptions",
	}, nil
}

func (s *RetrieveSubscriptionsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	subscriptions := pubsub.Elements().Child("subscriptions")
	nodeName := strings.Trim(subscriptions.Attributes().Get("node"), " ")
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	subscriptionsResult := xmpp.NewElementName("subscriptions")

	if nodeName == "" {
		nodeUserSubscriptionMap, err := repository.Repository().GetUserSubscriptions(*toJID.ToBareJID(), *fromJID.ToBareJID())
		if err != nil {
			log.Print(err.Error())
		}
		for tmpNodeName, item := range nodeUserSubscriptionMap {
			subscribers := item.GetSubscriptions()
			for _, item := range subscribers {
				elemSub := xmpp.NewElementName("subscription")
				elemSub.SetAttribute("node", tmpNodeName)
				elemSub.SetAttribute("jid", item.GetJid().String())
				elemSub.SetAttribute("subscription", item.GetSubscription().String())
				elemSub.SetAttribute("subid", item.GetSubid())
				subscriptionsResult.AppendElement(elemSub)
			}
		}
	} else {
		nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
		if nodeConfig == nil {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
		}

		subscriptionsResult.SetAttribute("node", nodeName)
		nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID.ToBareJID(), nodeName)
		subscribers := nodeSubscriptions.GetSubscriptions()
		for _, item := range subscribers {
			if item.GetJid().ToBareJID().String() != fromJID.ToBareJID().String() {
				continue
			}
			elemSub := xmpp.NewElementName("subscription")
			elemSub.SetAttribute("node", nodeName)
			elemSub.SetAttribute("jid", item.GetJid().String())
			elemSub.SetAttribute("subscription", item.GetSubscription().String())
			elemSub.SetAttribute("subid", item.GetSubid())

			subscriptionsResult.AppendElement(elemSub)
		}
	}

	pubSubResult.AppendElement(subscriptionsResult)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)
	return nil
}