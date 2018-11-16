package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"github.com/ortuman/jackal/component/pubsub/utils"
	"strings"
	"log"
)

type SubscribeNodeModule struct{}

func (s *SubscribeNodeModule) Name() string {
	return "SubscribeNodeModule"
}

func (s *SubscribeNodeModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleSubscribe := &base.ElementCriteria{}
	eleSubscribe.SetName("subscribe")

	elePubsub.AddCriteria(eleSubscribe)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *SubscribeNodeModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#manage-subscriptions",
		"http://jabber.org/protocol/pubsub#auto-subscribe",
		"http://jabber.org/protocol/pubsub#subscribe",
		// TODO "http://jabber.org/protocol/pubsub#subscription-notifications",
	}, nil
}

func (s *SubscribeNodeModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	subscribe := pubsub.Elements().Child("subscribe")
	nodeName := strings.Trim(subscribe.Attributes().Get("node"), " ")
	jidString := subscribe.Attributes().Get("jid")
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

	if nodeConfig.GetNodeAccessModel() == enums.AccessModelOpen {
		// TODO tigase : User blocked by domain
		if false { // replace [false] witch conditin of  [!Utils.isAllowedDomain(senderJid.getBareJID(), nodeConfig.getDomains())]
			errElem1 := xmpp.NewElementNamespace("text", "urn:ietf:params:xml:ns:xmpp-stanzas")
			errElem1.SetText("User blocked by domain")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
				errElem1,
			})
		}
	}

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	senderAffiliation := nodeAffiliations.GetSubscriberAffiliation(*fromJID.ToBareJID())

	if senderAffiliation.GetAffiliation() != enums.AffiliationOwner &&
	// TODO tiages : !this.config.isAdmin(senderJid)
		(jidErr != nil || jid.ToBareJID().String() != fromJID.ToBareJID().String()) {
		tmpErrElem := xmpp.NewElementNamespace("invalid-jid", "http://jabber.org/protocol/pubsub#errors")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest,
			[]xmpp.XElement{
				tmpErrElem,
			})
	}

	nodeSubscriptions := repository.Repository().GetNodeSubscriptions(*toJID.ToBareJID(), nodeName)
	subscription := nodeSubscriptions.GetSubscription(*jid.ToBareJID())

	// TODO 6.1.3.2 Presence Subscription Required
	// TODO 6.1.3.3 Not in Roster Group
	// TODO 6.1.3.4 Not on Whitelist
	// TODO 6.1.3.5 Payment Required
	// TODO 6.1.3.6 Anonymous NodeSubscriptions Not Allowed
	// TODO 6.1.3.9 NodeSubscriptions Not Supported
	// TODO 6.1.3.10 Node Has Moved

	if !senderAffiliation.GetAffiliation().IsSubscribe() {
		errElem := xmpp.NewElementNamespace("text", "urn:ietf:params:xml:ns:xmpp-stanzas")
		errElem.SetText("Not enough privileges to subscribe")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
			errElem,
		})
	}

	if subscription != enums.SubscriptionNone {
		if (subscription == enums.SubscriptionPending) && !(false || // TODO : replace [false] with admin check of [config.isAdmin(senderJid)]
			senderAffiliation.GetAffiliation() == enums.AffiliationOwner) {
			errElem1 := xmpp.NewElementNamespace("pending-subscription", "http://jabber.org/protocol/pubsub#errors")

			errElem2 := xmpp.NewElementNamespace("text", "urn:ietf:params:xml:ns:xmpp-stanzas")
			errElem2.SetText("Subscription is pending")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
				errElem1,
				errElem2,
			})
		}
	}

	accessModel := nodeConfig.GetNodeAccessModel()
	if (accessModel == enums.AccessModelWhitelist) &&
		(senderAffiliation.GetAffiliation() == enums.AffiliationNone || senderAffiliation.GetAffiliation() == enums.AffiliationOutcast) {
		errElem1 := xmpp.NewElementNamespace("closed-node", "http://jabber.org/protocol/pubsub#errors")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, []xmpp.XElement{
			errElem1,
		})
	}

	var results []xmpp.XElement
	var newSubscription enums.SubscriptionType
	nodeUserAffiliation := nodeAffiliations.GetSubscriberAffiliation(*jid.ToBareJID())
	affiliation := nodeUserAffiliation.GetAffiliation()

	if false || // TODO replace [false] with admin check of [(config.isAdmin(senderJid)]
		senderAffiliation.GetAffiliation() == enums.AffiliationOwner {
		newSubscription = enums.SubscriptionSubscribed
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationMember)
	} else if accessModel == enums.AccessModelOpen {
		newSubscription = enums.SubscriptionSubscribed
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationMember)
	} else if accessModel == enums.AccessModelAuthorize {
		newSubscription = enums.SubscriptionPending
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationNone)
	} else if accessModel == enums.AccessModelPresence {
		allowed := utils.DefaultPubSubLogic.HasSenderSubscription(*jid.ToBareJID(), nodeAffiliations, nodeSubscriptions)
		if !allowed {
			errElem1 := xmpp.NewElementNamespace("presence-subscription-required", "http://jabber.org/protocol/pubsub#errors")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrNotAuthorized, []xmpp.XElement{
				errElem1,
			})
		}
		newSubscription = enums.SubscriptionSubscribed
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationMember)
	} else if accessModel == enums.AccessModelRoster {
		allowed := utils.DefaultPubSubLogic.IsSenderInRosterGroup(*jid.ToBareJID(), nodeConfig, nodeAffiliations, nodeSubscriptions)
		if !allowed {
			errElem1 := xmpp.NewElementNamespace("not-in-roster-group", "http://jabber.org/protocol/pubsub#errors")
			return base.NewPubSubErrorStanza(packet, xmpp.ErrNotAuthorized, []xmpp.XElement{
				errElem1,
			})
		}
		newSubscription = enums.SubscriptionSubscribed
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationMember)
	} else if accessModel == enums.AccessModelWhitelist {
		newSubscription = enums.SubscriptionSubscribed
		affiliation = s.calculateNewOwnerAffiliation(affiliation, enums.AffiliationMember)
	} else {
		errElem1 := xmpp.NewElementNamespace("text", "urn:ietf:params:xml:ns:xmpp-stanzas")
		errElem1.SetText("AccessModel '" + accessModel.String() + "' is not implemented yet")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented, []xmpp.XElement{
			errElem1,
		})
	}

	subid := nodeSubscriptions.GetSubscriptionId(*jid.ToBareJID())
	var sendLastPublishedItem bool

	if subid == "" {
		subid = nodeSubscriptions.AddSubscriberJid(*jid.ToBareJID(), newSubscription)
		nodeAffiliations.AddAffiliation(*jid.ToBareJID(), affiliation)
		if (accessModel == enums.AccessModelAuthorize) &&
			(false || // replace [false] witch [!(this.config.isAdmin(senderJid)]
				senderAffiliation.GetAffiliation() == enums.AffiliationOwner) {

			pendingSubscriptionModule := GetModuleInstance("PendingSubscriptionModule")
			if pendingSubscriptionModule != nil {
				pendingSubscriptionModuleIns, ok := pendingSubscriptionModule.(*PendingSubscriptionModule)
				if ok {
					tmpArr := pendingSubscriptionModuleIns.SendAuthorizationRequest(nodeName, *toJID.ToBareJID(), subid, *jid, nodeAffiliations)
					results = append(results, tmpArr...)

				} else {
					log.Printf("PendingSubscriptionModule Instance not found")
				}
			} else {
				log.Printf("PendingSubscriptionModule Instance not found")
			}

		}
		// TODO sendLastPublishedItem
		// sendLastPublishedItem = nodeConfig.getSendLastPublishedItem() == SendLastPublishedItem.on_sub ||
		//						nodeConfig.getSendLastPublishedItem() == SendLastPublishedItem.on_sub_and_presence;
		sendLastPublishedItem = false
	} else {
		nodeSubscriptions.ChangeSubscription(*jid.ToBareJID(), newSubscription)
		nodeAffiliations.ChangeAffiliation(*jid.ToBareJID(), affiliation)
	}

	if nodeSubscriptions.SubscriptionsNeedsWriting() {
		repository.Repository().UpdateNodeSubscriptions(*toJID.ToBareJID(), nodeName, nodeSubscriptions)
	}
	if nodeAffiliations.AffiliationsNeedsWriting() {
		repository.Repository().UpdateNodeAffiliations(*toJID.ToBareJID(), nodeName, nodeAffiliations)
	}

	resultStanza.AppendElement(s.makeSubscription(nodeName, *jid, newSubscription, subid))
	results = append(results, resultStanza)

	for _, toSend := range results {
		stm.SendElement(toSend)
	}

	if sendLastPublishedItem {
		// TODO
		// publishItemModule.publishLastItem(serviceJid, nodeConfig, JID.jidInstance(jid))
	}
	return nil
}

func (s *SubscribeNodeModule) calculateNewOwnerAffiliation(ownerAffiliation, newAffiliation enums.AffiliationType) enums.AffiliationType {
	if ownerAffiliation.Weight() > newAffiliation.Weight() {
		return ownerAffiliation
	}
	return newAffiliation
}

func (s *SubscribeNodeModule)makeSubscription(nodeName string, subscriberJid jid.JID, newSubscription enums.SubscriptionType, subid string) *xmpp.Element {
	resPubSub := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	resSubscription := xmpp.NewElementName("subscription")
	resSubscription.SetAttribute("node", nodeName)
	resSubscription.SetAttribute("jid", subscriberJid.String())
	resSubscription.SetAttribute("subscription", newSubscription.String())
	resSubscription.SetAttribute("subid", subid)
	resPubSub.AppendElement(resSubscription)
	return resPubSub
}