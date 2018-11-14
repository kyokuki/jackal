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

type PurgeItemsModule struct{}

func (s *PurgeItemsModule) Name() string {
	return "PurgeItemsModule"
}

func (s *PurgeItemsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	elePurge := &base.ElementCriteria{}
	elePurge.SetName("purge")

	elePubsub.AddCriteria(elePurge)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *PurgeItemsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#purge-nodes",
	}, nil
}

func (s *PurgeItemsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	purge := pubsub.Elements().Child("purge")
	nodeName := strings.Trim(purge.Attributes().Get("node"), " ")
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
			errElem.SetAttribute("feature", "purge-items")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
				[]xmpp.XElement{
					errElem,
				})
		}
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID)

	if !senderAffiliation.GetAffiliation().IsPurgeNode() {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
	}

	leafNodeConfig, ok := nodeConfig.(*base.LeafNodeConfig)
	if !ok {
		errElem := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		errElem.SetAttribute("feature", "purge-items")
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

	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID.ToBareJID(), nodeName)
	itemsToDelete, _ := repository.Repository().GetItemIds(*toJID.ToBareJID(), nodeName)

	publishModule := GetModuleInstance("PublishItemModule")
	if publishModule != nil {
		publishModuleIns, ok := publishModule.(*PublishItemModule)
		if ok {
			elemPurge := xmpp.NewElementName("purge")
			elemPurge.SetAttribute("node", nodeName)

			publishModuleIns.SendNotificationsByItemElement(elemPurge, *toJID.ToBareJID(), nodeName, nil, nodeConfig, nodeAffiliations, nodeSubscriptions)
			log.Printf("Purging node " + nodeName)
		} else {
			log.Printf("PublishItemModule Instance not found")
		}
	} else {
		log.Printf("PublishItemModule Instance not found")
	}

	for _, itemId := range itemsToDelete {
		repository.Repository().DeleteItem(*toJID.ToBareJID(), nodeName, itemId)
	}

	stm.SendElement(resultStanza)
	return nil
}
