package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"strings"
)

type UnsubscribeNodeModule struct{}

func (s *UnsubscribeNodeModule) Name() string {
	return "UnsubscribeNodeModule"
}

func (s *UnsubscribeNodeModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleUnsubscribe := &base.ElementCriteria{}
	eleUnsubscribe.SetName("unsubscribe")

	elePubsub.AddCriteria(eleUnsubscribe)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *UnsubscribeNodeModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{}, nil
}

func (s *UnsubscribeNodeModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	unsubscribe := pubsub.Elements().Child("unsubscribe")
	nodeName := strings.Trim(unsubscribe.Attributes().Get("node"), " ")
	jidString := unsubscribe.Attributes().Get("jid")
	subid := strings.Trim(unsubscribe.Attributes().Get("subid"), " ")
	jid, jidErr := jid.NewWithString(jidString, false)
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	if nodeName == "" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("nodeid-required", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID.ToBareJID())
	nodeUserAffiliation := nodeAffiliations.GetSubscriberAffiliation(*jid.ToBareJID())
	affiliation := nodeUserAffiliation.GetAffiliation()

	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID.ToBareJID(), nodeName)
	subscription := nodeSubscriptions.GetSubscription(*jid.ToBareJID())

	if (senderAffiliation.GetAffiliation() != enums.AffiliationOwner) && (jidErr != nil || jid.ToBareJID().String() != fromJID.ToBareJID().String()) {
		tmpErrElem := xmpp.NewElementNamespace("invalid-jid", "http://jabber.org/protocol/pubsub#errors")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				tmpErrElem,
			})
	}

	if affiliation == enums.AffiliationOutcast {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
	}

	if subid != "" {
		selectSubid := nodeSubscriptions.GetSubscriptionId(*jid.ToBareJID())
		if subid != selectSubid {
			errElem1 := xmpp.NewElementNamespace("not-acceptable", "urn:ietf:params:xml:ns:xmpp-stanzas")
			errElem2 := xmpp.NewElementNamespace("invalid-subid", "http://jabber.org/protocol/pubsub#errors")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
				[]xmpp.XElement{
					errElem1,
					errElem2,
				})
		}
		subid = selectSubid
	}

	if subscription == enums.SubscriptionNone {
		errElem1 := xmpp.NewElementNamespace("unexpected-request", "urn:ietf:params:xml:ns:xmpp-stanzas")
		errElem2 := xmpp.NewElementNamespace("not-subscribed", "http://jabber.org/protocol/pubsub#errors")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				errElem1,
				errElem2,
			})
	}

	nodeSubscriptions.ChangeSubscription(*jid.ToBareJID(), enums.SubscriptionNone)
	if nodeSubscriptions.SubscriptionsNeedsWriting() {
		repository.Repository().UpdateNodeSubscriptions(*toJID.ToBareJID(), nodeName, nodeSubscriptions)
	}

	resultStanza.AppendElement(s.makeSubscription(nodeName, *jid, enums.SubscriptionNone, subid))
	stm.SendElement(resultStanza)
	return nil
}


func (s *UnsubscribeNodeModule)makeSubscription(nodeName string, subscriberJid jid.JID, newSubscription enums.SubscriptionType, subid string) *xmpp.Element {
	resPubSub := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	resSubscription := xmpp.NewElementName("subscription")
	resSubscription.SetAttribute("node", nodeName)
	resSubscription.SetAttribute("jid", subscriberJid.String())
	resSubscription.SetAttribute("subscription", newSubscription.String())
	resSubscription.SetAttribute("subid", subid)
	resPubSub.AppendElement(resSubscription)
	return resPubSub
}