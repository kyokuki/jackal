package utils

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type privatePubSubLogic struct {

}

var DefaultPubSubLogic privatePubSubLogic

func (psl *privatePubSubLogic) HasSenderSubscription(jid jid.JID, nodeAffiations *cached.NodeAffiliations, nodeSubscriptions *cached.NodeSubscriptions) bool {
	// TODO
	return false
}

func (psl *privatePubSubLogic) IsSenderInRosterGroup(jid jid.JID, nodeConfig base.AbstractNodeConfig, nodeAffiations *cached.NodeAffiliations, nodeSubscriptions *cached.NodeSubscriptions) bool {
	// TODO
	return false
}

func (psl *privatePubSubLogic) CheckAccessPermission(packet xmpp.Stanza, serviceJid jid.JID, nodeName string, senderJid jid.JID) (bool, *base.PubSubError) {
	// TODO : check whether is admin
	if false {
		return true, nil
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(serviceJid, nodeName)
	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(serviceJid, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(senderJid)
	senderSubscription := nodeSubscriptions.GetSubscription(senderJid)

	nodeConfig := repository.Repository().GetNodeConfig(serviceJid, nodeName)
	if nodeConfig == nil {
		return false, base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	}

	// TODO : check whether allow domain
	if nodeConfig.GetNodeAccessModel() == enums.AccessModelOpen {
		return true, nil
	}

	if senderAffiliation.GetAffiliation() == enums.AffiliationOutcast {
		return false, base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)
	}

	nodeAccessModel := nodeConfig.GetNodeAccessModel()
	if nodeAccessModel == enums.AccessModelWhitelist && !senderAffiliation.GetAffiliation().IsRetrieveItem() {
		return false, base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
			xmpp.NewElementNamespace("closed-node", "http://jabber.org/protocol/pubsub#errors"),
		})
	} else if (nodeAccessModel == enums.AccessModelAuthorize) && (senderSubscription != enums.SubscriptionSubscribed || senderAffiliation.GetAffiliation().IsRetrieveItem()) {
		return false, base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
			xmpp.NewElementNamespace("not-subscribed", "http://jabber.org/protocol/pubsub#errors"),
		})
	} else if nodeAccessModel == enums.AccessModelPresence {
		allowd := psl.HasSenderSubscription(senderJid, nodeAffiliations, nodeSubscriptions)
		if !allowd {
			return false, base.NewPubSubErrorStanza(packet, xmpp.ErrNotAuthorized, []xmpp.XElement{
				xmpp.NewElementNamespace("presence-subscription-required", "http://jabber.org/protocol/pubsub#errors"),
			})
		}
	} else if  nodeAccessModel == enums.AccessModelRoster {
		allowd := psl.IsSenderInRosterGroup(senderJid, nodeConfig, nodeAffiliations, nodeSubscriptions)
		if !allowd {
			return false, base.NewPubSubErrorStanza(packet, xmpp.ErrNotAuthorized, []xmpp.XElement{
				xmpp.NewElementNamespace("not-in-roster-group", "http://jabber.org/protocol/pubsub#errors"),
			})
		}
	}

	return false, base.NewPubSubErrorStanza(packet, xmpp.ErrNotAuthorized, []xmpp.XElement{})
}

func (psl *privatePubSubLogic) PrepareNotificationMessage(fromJid jid.JID, toJid jid.JID, id string, itemToSend xmpp.XElement, headers map[string]string) xmpp.XElement {
	elemMessage := xmpp.NewElementNamespace("message", "jabber:client")
	elemMessage.SetAttribute("from", fromJid.String())
	elemMessage.SetAttribute("to", toJid.String())
	elemMessage.SetAttribute("id", id)
	elemEvent := xmpp.NewElementNamespace("event", "http://jabber.org/protocol/pubsub#event")
	elemEvent.AppendElement(itemToSend)
	elemMessage.AppendElement(elemEvent)
	if len(headers) > 0 {
		elemHeader := xmpp.NewElementNamespace("headers", "http://jabber.org/protocol/shim")
		for iKey, iVal := range headers {
			h := xmpp.NewElementName("header")
			h.SetAttribute("name", iKey)
			h.SetText(iVal)
			elemHeader.AppendElement(h)
		}
		elemMessage.AppendElement(elemHeader)
	}

	return elemMessage
}