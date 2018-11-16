package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"log"
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
)

type PendingSubscriptionModule struct{}

func (s *PendingSubscriptionModule) Name() string {
	return "PendingSubscriptionModule"
}

func (s *PendingSubscriptionModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("message")
	eleX := &base.ElementCriteria{}
	eleX.SetName("x").AddAttr("xmlns", "jabber:x:data").AddAttr("type", "submit")
	eleField := &base.ElementCriteria{}
	eleField.SetName("field").AddAttr("var", "FORM_TYPE")
	eleValue := &base.ElementCriteria{}
	eleValue.SetName("value").SetCDATA("http://jabber.org/protocol/pubsub#subscribe_authorization")

	eleField.AddCriteria(eleValue)
	eleX.AddCriteria(eleX)
	eleCrit.AddCriteria(eleX)
	return eleCrit
}

func (s *PendingSubscriptionModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#get-pending",
	}, nil
}

func (s *PendingSubscriptionModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	id := packet.ID()
	elemX := packet.Elements().ChildNamespace("x", "jabber:x:data")
	formX, err := xep0004.NewFormFromElement(elemX)
	if err != nil {
		log.Printf("NewFormFromElement error:" + err.Error())
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}

	_, fieldSubId := formX.Field("pubsub#subid")
	valueSubId := fieldSubId.GetAsString()
	_, fieldNode := formX.Field("pubsub#node")
	valueNode := fieldNode.GetAsString()
	_, fieldSubscriberJid := formX.Field("pubsub#subscriber_jid")
	valueSubscriberJidString := fieldSubscriberJid.GetAsString()
	valueSubscriberJid, err := jid.NewWithString(valueSubscriberJidString, false)
	if err != nil {
		log.Printf("jid.NewWithString of [" + valueSubscriberJidString + "] error:" + err.Error())
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}
	_, fieldAllow := formX.Field("pubsub#allow")
	valueAllowString := fieldAllow.GetAsBooleanString()

	if valueAllowString == "" {
		return nil
	}


	if valueNode == "" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("node-required", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, valueNode)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, valueNode)
	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID, valueNode)

	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID)
	if senderAffiliation.GetAffiliation() != enums.AffiliationOwner {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{})
	}

	userSubId := nodeSubscriptions.GetSubscriptionId(*valueSubscriberJid)
	if valueSubId != "" && valueSubId != userSubId {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrNotAcceptable,
			[]xmpp.XElement{
				xmpp.NewElementNamespace("invalid-subid", "http://jabber.org/protocol/pubsub#errors"),
			})
	}

	subscription := nodeSubscriptions.GetSubscription(*valueSubscriberJid)
	if subscription != enums.SubscriptionPending {
		return nil
	}

	affiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID.ToBareJID()).GetAffiliation()

	if valueAllowString == "true" {
		subscription = enums.SubscriptionSubscribed
		affiliation = enums.AffiliationMember
		nodeSubscriptions.ChangeSubscription(*valueSubscriberJid, subscription)
		nodeAffiliations.ChangeAffiliation(*valueSubscriberJid, affiliation)
	} else {
		subscription = enums.SubscriptionNone
		nodeSubscriptions.ChangeSubscription(*valueSubscriberJid, subscription)
	}

	if nodeSubscriptions.SubscriptionsNeedsWriting() {
		repository.Repository().UpdateNodeSubscriptions(*toJID, valueNode, nodeSubscriptions)
	}
	if nodeAffiliations.AffiliationsNeedsWriting() {
		repository.Repository().UpdateNodeAffiliations(*toJID, valueNode, nodeAffiliations)
	}


	eleMessage := xmpp.NewElementNamespace(packet.Name(), "jabber:client")
	eleMessage.SetTo(valueSubscriberJid.String())
	eleMessage.SetFrom(toJID.String())
	eleMessage.SetAttribute("id", id)

	subscribeNodeModule := GetModuleInstance("SubscribeNodeModule")
	if subscribeNodeModule != nil {
		subscribeNodeModuleIns, ok := subscribeNodeModule.(*SubscribeNodeModule)
		if ok {
			elemSub := subscribeNodeModuleIns.makeSubscription(valueNode, *valueSubscriberJid, subscription, "")
			eleMessage.AppendElement(elemSub)
			stm.SendElement(eleMessage)
		} else {
			log.Printf("SubscribeNodeModule Instance not found")
		}
	} else {
		log.Printf("SubscribeNodeModule Instance not found")
	}

	return nil
}


func (s *PendingSubscriptionModule) SendAuthorizationRequest(nodeName string, fromJid jid.JID, subId string, subscriberJid jid.JID, nodeAffiliations cached.NodeAffiliations) ([]xmpp.XElement) {
	formX := xep0004.DataForm{}
	formX.Type = "form"
	formX.Title = "PubSub subscriber request"
	formX.Instructions = "To approve this entity's subscription request, click the OK button. To deny the request, click the cancel button."

	formX.AddField(xep0004.NewFieldHidden("FORM_TYPE", "http://jabber.org/protocol/pubsub#subscribe_authorization"))
	formX.AddField(xep0004.NewFieldHidden("pubsub#subid", subId))
	formX.AddField(xep0004.NewFieldTextSingle("pubsub#node", nodeName, "Node ID"))
	formX.AddField(xep0004.NewFieldTextSingle("pubsub#node", nodeName, "Node ID"))
	formX.AddField(xep0004.NewFieldJidSingle("pubsub#subscriber_jid", subscriberJid.String(), "UsersSubscription Address"))
	formX.AddField(xep0004.NewFieldBool("pubsub#allow", false, "Allow this JID to subscribe to this pubsub node?"))

	var result []xmpp.XElement
	affiliations := nodeAffiliations.GetAffiliations()
	for _, userAff := range affiliations {
		if userAff.GetAffiliation() == enums.AffiliationOwner {
			eleMessage := xmpp.NewElementNamespace("message", "jabber:client")
			eleMessage.SetTo(userAff.GetJid().String())
			eleMessage.SetFrom(fromJid.ToBareJID().String())
			eleMessage.SetAttribute("id", "")
			eleMessage.AppendElement(formX.Element())

			result = append(result, eleMessage)
		}
	}

	return result
}