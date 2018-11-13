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
	"github.com/satori/go.uuid"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"github.com/ortuman/jackal/component/pubsub/utils"
	"log"
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

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID)
	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID, nodeName)
	senderSubscription := nodeSubscriptions.GetSubscription(*fromJID.ToBareJID())

	publisherModel := nodeConfig.GetPublisherModel()

	if !senderAffiliation.GetAffiliation().IsPublishItem() {
		if publisherModel == enums.PublisherModelPublishers {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
		}

		if publisherModel == enums.PublisherModelSubscribers && senderSubscription != enums.SubscriptionSubscribed {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
		}
	}

	itemsToSend := s.makeItemsToSend(publish)
	leafNodeConfig, ok := nodeConfig.(*base.LeafNodeConfig)
	if !ok {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	if leafNodeConfig.IsPersistItem() {
		for i := 0; i < len(itemsToSend); i++ {
			loopItem := itemsToSend[i].(*xmpp.Element)
			itemId := loopItem.Attributes().Get("id")
			if itemId == "" {
				u1 := uuid.Must(uuid.NewV1())
				itemId = strings.Replace(u1.String(), "-", "", -1)
				loopItem.SetAttribute("id", itemId)
			}

			itemElem := xmpp.NewElementName("item")
			itemElem.SetAttribute("id", itemId)
			publishResult.AppendElement(itemElem)
		}
	}

	pubSubResult.AppendElement(publishResult)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)

	s.doPublishItems(*toJID, nodeName, leafNodeConfig, *fromJID.ToBareJID(), itemsToSend)
	return nil
}

func (s *PublishItemModule) makeItemsToSend(publishElem xmpp.XElement) []xmpp.XElement {
	var itemArr []xmpp.XElement
	for _, item := range publishElem.Elements().All() {
		if "item" != item.Name() {
			continue
		}

		expireAttr := item.Attributes().Get("expire-at")
		if expireAttr != "" {

		}

		itemArr = append(itemArr, item)
	}
	return itemArr
}

func (s *PublishItemModule) doPublishItems(serviceJID jid.JID, nodeName string, leafNodeConfig *base.LeafNodeConfig, publisherJID jid.JID, itemsToSend []xmpp.XElement) error {
	if leafNodeConfig.IsPersistItem() {
		for _, loopItem := range itemsToSend {
			itemId := loopItem.Attributes().Get("id")
			err := repository.Repository().WriteItem(serviceJID, nodeName, itemId, publisherJID, loopItem)
			if err != nil {
				log.Printf(s.Name() + " - Error processing publish packet - " + err.Error())
				return err
			}
		}

		if leafNodeConfig.MaxItems() > 0 {
			// TODO trim items
		}
	}

	return s.SendNotificationsByItemsSlice(serviceJID, nodeName, itemsToSend)
}

func (s *PublishItemModule) SendNotificationsByItemsSlice(serviceJID jid.JID, nodeName string, itemsToSend []xmpp.XElement) error {
	nodeAffiliations := repository.Repository().GetNodeAffiliations(serviceJID, nodeName)
	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(serviceJID, nodeName)

	items := xmpp.NewElementName("items")
	items.SetAttribute("node", nodeName)
	for _, loopItem := range itemsToSend {
		items.AppendElement(loopItem)
	}

	return s.SendNotificationsByItemElement(items, serviceJID, nodeName, nil,
		repository.Repository().GetNodeConfig(serviceJID, nodeName), nodeAffiliations, nodeSubscriptions)
}

func (s *PublishItemModule) SendNotificationsByItemElement(
	itemToSend xmpp.XElement,
	fromJid jid.JID,
	nodeName string,
	headers map[string]string,
	nodeConfig base.AbstractNodeConfig,
	nodeAfffiliations *cached.NodeAffiliations,
	nodeSubscriptions *cached.NodeSubscriptions,
) error {

	var subscriberJids []jid.JID
	subscriberJids = s.GetActiveSubscribers(nodeConfig, nodeAfffiliations, nodeSubscriptions)

	// TODO IsDeliverPresenceBased
	if nodeConfig.IsDeliverPresenceBased() {

	}

	return s.SendNotificationsBySubscribers(subscriberJids, itemToSend, fromJid, nodeConfig, nodeName, headers)
}

func (s *PublishItemModule) SendNotificationsBySubscribers(
	subscribers []jid.JID,
	itemToSend xmpp.XElement,
	fromJid jid.JID,
	nodeConfig base.AbstractNodeConfig,
	nodeName string,
	headers map[string]string,
) error {
	for _, jid := range subscribers {
		eleMessage := utils.DefaultPubSubLogic.PrepareNotificationMessage(fromJid, jid, "", itemToSend, headers)
		packet := PacketInstance(eleMessage, fromJid, jid)
		GetStreamC2S().SendElement(packet)
	}
	return nil
}

func (s *PublishItemModule) GetActiveSubscribers(nodeConfig base.AbstractNodeConfig, affiliations *cached.NodeAffiliations, subscriptions *cached.NodeSubscriptions) []jid.JID {
	activeUserSubscribers := subscriptions.GetSubscriptionsForPublish()
	var resultJIDs []jid.JID
	for _, userSub := range activeUserSubscribers {
		tmpJid := *userSub.GetJid()
		affiliation := affiliations.GetSubscriberAffiliation(tmpJid)

		if affiliation.GetAffiliation() != enums.AffiliationOutcast {
			subscription := subscriptions.GetSubscription(tmpJid)
			if subscription == enums.SubscriptionSubscribed {
				resultJIDs = append(resultJIDs, tmpJid)
			}
		}
	}

	return resultJIDs
}
