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
	"log"
)

type RetractItemModule struct{}

func (s *RetractItemModule) Name() string {
	return "RetractItemModule"
}

func (s *RetractItemModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleRetract := &base.ElementCriteria{}
	eleRetract.SetName("retract")

	elePubsub.AddCriteria(eleRetract)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *RetractItemModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#retract-items",
	}, nil
}

func (s *RetractItemModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	retract := pubsub.Elements().Child("retract")
	nodeName := strings.Trim(retract.Attributes().Get("node"), " ")
	id := packet.ID()
	//stanzaType := strings.ToLower(packet.Attributes().Get("type"))

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

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
			errElem := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
			errElem.SetAttribute("feature", "retract-items")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
				[]xmpp.XElement{
					errElem,
				})
		}
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID)

	if !senderAffiliation.GetAffiliation().IsDeleteItem() {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
	}

	leafNodeConfig, ok := nodeConfig.(*base.LeafNodeConfig)
	if !ok {
		errElem := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		errElem.SetAttribute("feature", "retract-items")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
			[]xmpp.XElement{
				errElem,
			})
	}
	if !leafNodeConfig.IsPersistItem() {
		errElem := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		errElem.SetAttribute("feature", "persistent-items")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
			[]xmpp.XElement{
				errElem,
			})
	}

	var itemsToDelete []string
	for _, loopItem := range retract.Elements().All() {
		if "item" != loopItem.Name() {
			continue
		}

		id := loopItem.Attributes().Get("id")
		if id != "" {
			itemsToDelete = append(itemsToDelete, id)
		} else {
			errElem := xmpp.NewElementNamespace("item-required", "http://jabber.org/protocol/pubsub#errors")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
				[]xmpp.XElement{
					errElem,
				})
		}
	}
	if len(itemsToDelete) == 0 {
		errElem := xmpp.NewElementNamespace("item-required", "http://jabber.org/protocol/pubsub#errors")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				errElem,
			})
	}

	var itemsToSend []xmpp.XElement
	for _, itemId := range itemsToDelete {
		repository.Repository().DeleteItem(*toJID.ToBareJID(), nodeName, itemId)

		elemNotification := xmpp.NewElementName("retract")
		elemNotification.SetAttribute("id", itemId)

		itemsToSend = append(itemsToSend, elemNotification)
	}

	publishModule := GetModuleInstance("PublishItemModule")
	if publishModule != nil {
		publishModuleIns, ok := publishModule.(*PublishItemModule)
		if ok {
			publishModuleIns.SendNotificationsByItemsSlice(*toJID.ToBareJID(), nodeName, itemsToSend)
		} else {
			log.Printf("PublishItemModule Instance not found")
		}
	} else {
		log.Printf("PublishItemModule Instance not found")
	}

	stm.SendElement(resultStanza)
	return nil
}
