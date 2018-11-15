package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"strings"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type ManageSubscriptionsModule struct{}

func (s *ManageSubscriptionsModule) Name() string {
	return "ManageSubscriptionsModule"
}

func (s *ManageSubscriptionsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	eleSubscriptions := &base.ElementCriteria{}
	eleSubscriptions.SetName("subscriptions")

	elePubsub.AddCriteria(eleSubscriptions)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *ManageSubscriptionsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#manage-subscriptions",
	}, nil
}

func (s *ManageSubscriptionsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	subscriptions := pubsub.Elements().Child("subscriptions")
	nodeName := strings.Trim(subscriptions.Attributes().Get("node"), " ")
	id := packet.ID()
	stanzaType := packet.Attributes().Get("type")

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	subscriptionsResult := xmpp.NewElementName("subscriptions")

	if nodeName == "" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("node-required", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	}

	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID.ToBareJID(), nodeName)
	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID.ToBareJID(), nodeName)
	senderJid := *fromJID.ToBareJID()

	allowed := s.checkPrivileges(packet, stanzaType, senderJid, nodeConfig, nodeAffiliations, nodeSubscriptions)
	if !allowed {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)
	}

	if strings.ToLower(stanzaType) == "get" {
		return s.processGet(packet, nodeName, nodeSubscriptions)
	} else if strings.ToLower(stanzaType) == "set" {
		return s.processSet(packet, nodeSubscriptions)
	} else {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}

	pubSubResult.AppendElement(subscriptionsResult)
	resultStanza.AppendElement(pubSubResult)
	//stm.SendElement(resultStanza)
	return nil
}

func (s *ManageSubscriptionsModule) processGet(packet xmpp.Stanza, nodeName string, nodeSubscriptions *cached.NodeSubscriptions) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	subscriptionsResult := xmpp.NewElementName("subscriptions")
	subscriptionsResult.SetAttribute("node", nodeName)

	subscribers := nodeSubscriptions.GetSubscriptions()
	for _, item := range subscribers {
		elemSub := xmpp.NewElementName("subscription")
		elemSub.SetAttribute("node", nodeName)
		elemSub.SetAttribute("jid", item.GetJid().String())
		elemSub.SetAttribute("subscription", item.GetSubscription().String())
		elemSub.SetAttribute("subid", item.GetSubid())

		subscriptionsResult.AppendElement(elemSub)
	}

	pubSubResult.AppendElement(subscriptionsResult)
	resultStanza.AppendElement(pubSubResult)
	GetStreamC2S().SendElement(resultStanza)
	return nil
}

func (s *ManageSubscriptionsModule) processSet(packet xmpp.Stanza, nodeSubscriptions *cached.NodeSubscriptions) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	subscriptions := pubsub.Elements().Child("subscriptions")
	nodeName := strings.Trim(subscriptions.Attributes().Get("node"), " ")
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	for _, subItemElem := range subscriptions.Elements().All() {
		if "subscription" != subItemElem.Name() {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
		}
	}

	changedSubscriptions := make(map[jid.JID]enums.SubscriptionType)
	for _, subItemElem := range subscriptions.Elements().All() {
		strSubscription := subItemElem.Attributes().Get("subscription")
		strJid := subItemElem.Attributes().Get("jid")
		jid, _ := jid.NewWithString(strJid, false)

		if strSubscription == "" {
			continue
		}

		newSubscription := enums.NewSubscriptionValue(strSubscription)
		oldSubscription := nodeSubscriptions.GetSubscription(*jid.ToBareJID())
		if (oldSubscription == enums.SubscriptionNone) && (newSubscription != enums.SubscriptionNone) {
			nodeSubscriptions.AddSubscriberJid(*jid.ToBareJID(), newSubscription)
			changedSubscriptions[*jid.ToBareJID()] = newSubscription
		} else {
			nodeSubscriptions.ChangeSubscription(*jid.ToBareJID(), newSubscription)
			changedSubscriptions[*jid.ToBareJID()] = newSubscription
		}
	}

	if nodeSubscriptions.SubscriptionsNeedsWriting() {
		repository.Repository().UpdateNodeSubscriptions(*toJID.ToBareJID(), nodeName, nodeSubscriptions)
	}

	// TODO Notify Changed Subscription Affiliation by jackal's configuration
	//for mapJid, mapSub := range changedSubscriptions {
	//	// do notify
	//}

	GetStreamC2S().SendElement(resultStanza)
	return nil
}

func (s *ManageSubscriptionsModule) checkPrivileges(packet xmpp.Stanza, stanzaType string, senderJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeAffiliations *cached.NodeAffiliations, nodeSubscriptions *cached.NodeSubscriptions) bool {
	allowed := false
	if !allowed {
		if strings.ToLower(stanzaType) == "get" {
			senderSubscription := nodeSubscriptions.GetSubscription(senderJid)
			if senderSubscription == enums.SubscriptionSubscribed {
				allowed = true
			}
		}
	}

	if !allowed {
		senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(senderJid)
		if senderAffiliation.GetAffiliation() == enums.AffiliationOwner {
			allowed = true
		}
	}

	if !allowed {
		// TODO
		//if this.config.isAdmin(senderJid) == true {
		//	allowed = true
		//}
	}
	return allowed
}
