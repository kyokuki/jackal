package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"strings"
	"github.com/ortuman/jackal/component/pubsub/utils"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"log"
	"strconv"
)

type RetrieveItemsModule struct{}

func (s *RetrieveItemsModule) Name() string {
	return "RetrieveItemsModule"
}

func (s *RetrieveItemsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleitems := &base.ElementCriteria{}
	eleitems.SetName("items")

	elePubsub.AddCriteria(eleitems)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *RetrieveItemsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#retrieve-items",
	}, nil
}

func (s *RetrieveItemsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	items := pubsub.Elements().Child("items")
	nodeName := strings.Trim(items.Attributes().Get("node"), " ")
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

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	itemsResult := xmpp.NewElementName("items")
	itemsResult.SetAttribute("node", nodeName)

	allowed, falseElement := utils.DefaultPubSubLogic.CheckAccessPermission(packet, *toJID.ToBareJID(), nodeName, *fromJID.ToBareJID())
	if !allowed {
		if falseElement != nil {
			return falseElement
		}
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}

	var requestedIds []string
	for _, loopItem := range items.Elements().All() {
		strId := loopItem.Attributes().Get("id")
		if "item" != loopItem.Name() || strId == "" {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
		}
		requestedIds = append(requestedIds, strId)
	}

	if len(requestedIds) > 0 {
		for _, loopId := range requestedIds {
			itemMeta, itemErr := repository.Repository().GetItem(*toJID.ToBareJID(), nodeName, loopId)
			if itemErr != nil {
				log.Printf("pubsub node item retrieve error : %s", itemErr.Error())
				continue
			}

			p := xmpp.NewParser(strings.NewReader(itemMeta.Data), xmpp.DefaultMode, 0)
			payload, payloadErr := p.ParseElement()
			if payloadErr != nil {
				log.Printf("pubsub node item's payload error : %s", payloadErr.Error())
				continue
			}
			itemsResult.AppendElement(payload)
		}
	} else {
		// TODO
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented, []xmpp.XElement{})

		maxItemsStr := items.Attributes().Get("max_items")
		var maxItemsCnt int
		rsmGet := pubsub.Elements().ChildNamespace("set", "http://jabber.org/protocol/rsm")
		if rsmGet != nil {

		} else {
			if maxItemsStr != "" {
				maxItemsCnt, _ = strconv.Atoi(maxItemsStr)
			}
		}

		itemArr, err := repository.Repository().QueryItems(*toJID.ToBareJID(), nodeName, int64(maxItemsCnt))
		if err != nil {
			log.Printf("pubsub node item retrieve error : %s", err.Error())
		}

		for _, loopItem := range itemArr {
			p := xmpp.NewParser(strings.NewReader(loopItem.Data), xmpp.DefaultMode, 0)
			payload, payloadErr := p.ParseElement()
			if payloadErr != nil {
				log.Printf("pubsub node item's payload error : %s", payloadErr.Error())
				continue
			}
			itemsResult.AppendElement(payload)
		}
	}

	pubSubResult.AppendElement(itemsResult)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)
	return nil
}
