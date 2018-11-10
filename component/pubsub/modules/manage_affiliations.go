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

type ManageAffiliationsModule struct{}

func (s *ManageAffiliationsModule) Name() string {
	return "ManageAffiliationsModule"
}

func (s *ManageAffiliationsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	eleAffiliations := &base.ElementCriteria{}
	eleAffiliations.SetName("affiliations")

	elePubsub.AddCriteria(eleAffiliations)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *ManageAffiliationsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#manage-affiliations",
	}, nil
}

func (s *ManageAffiliationsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	affiliations := pubsub.Elements().Child("affiliations")
	nodeName := strings.Trim(affiliations.Attributes().Get("node"), " ")
	id := packet.ID()
	stanzaType := strings.ToLower(packet.Attributes().Get("type"))

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	affiliationsResult := xmpp.NewElementName("affiliations")

	if stanzaType != "set" && stanzaType != "get" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}

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

	nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID.ToBareJID(), nodeName)
	senderJid := *fromJID.ToBareJID()

	allowed := s.checkPrivileges(packet, stanzaType, senderJid, nodeConfig, nodeAffiliations)
	if !allowed {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)
	}

	if strings.ToLower(stanzaType) == "get" {
		return s.processGet(packet, stm, nodeName, nodeAffiliations)
	} else if strings.ToLower(stanzaType) == "set" {
		return s.processSet(packet, stm, nodeAffiliations)
	} else {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
	}

	pubSubResult.AppendElement(affiliationsResult)
	resultStanza.AppendElement(pubSubResult)
	//stm.SendElement(resultStanza)
	return nil
}

func (s *ManageAffiliationsModule) processGet(packet xmpp.Stanza, stm stream.C2S, nodeName string, nodeAffiliations *cached.NodeAffiliations) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	affiliationsResult := xmpp.NewElementName("affiliations")
	affiliationsResult.SetAttribute("node", nodeName)

	subAffers := nodeAffiliations.GetAffiliations()
	for _, item := range subAffers {
		elemAff := xmpp.NewElementName("affiliation")
		elemAff.SetAttribute("node", nodeName)
		elemAff.SetAttribute("jid", item.GetJid().String())
		elemAff.SetAttribute("subscription", item.GetAffiliation().String())

		affiliationsResult.AppendElement(elemAff)
	}

	pubSubResult.AppendElement(affiliationsResult)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)
	return nil
}

func (s *ManageAffiliationsModule) processSet(packet xmpp.Stanza, stm stream.C2S, nodeAffiliations *cached.NodeAffiliations) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	affiliations := pubsub.Elements().Child("affiliations")
	nodeName := strings.Trim(affiliations.Attributes().Get("node"), " ")
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	for _, subItemElem := range affiliations.Elements().All() {
		if "affiliation" != subItemElem.Name() {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, []xmpp.XElement{})
		}
	}

	changedAffiliations := make(map[jid.JID]enums.AffiliationType)
	for _, subItemElem := range affiliations.Elements().All() {
		strAffiliations := subItemElem.Attributes().Get("affiliation")
		strJid := subItemElem.Attributes().Get("jid")
		jid, _ := jid.NewWithString(strJid, false)

		if strAffiliations == "" {
			continue
		}

		newAffiliation := enums.NewAffiliationValue(strAffiliations)
		subscriberAffiliation := nodeAffiliations.GetSubscriberAffiliation(*jid.ToBareJID())
		oldAffiliation := subscriberAffiliation.GetAffiliation()
		if (oldAffiliation == enums.AffiliationNone) && (newAffiliation != enums.AffiliationNone) {
			nodeAffiliations.AddAffiliation(*jid.ToBareJID(), newAffiliation)
			changedAffiliations[*jid.ToBareJID()] = newAffiliation
		} else {
			nodeAffiliations.ChangeAffiliation(*jid.ToBareJID(), newAffiliation)
			changedAffiliations[*jid.ToBareJID()] = newAffiliation
		}
	}

	if nodeAffiliations.AffiliationsNeedsWriting() {
		repository.Repository().UpdateNodeAffiliations(*toJID.ToBareJID(), nodeName, nodeAffiliations)
	}

	// TODO Notify Changed Subscription Affiliation by jackal's configuration
	//for mapJid, mapSub := range changedSubscriptions {
	//	// do notify
	//}

	stm.SendElement(resultStanza)
	return nil
}

func (s *ManageAffiliationsModule) checkPrivileges(packet xmpp.Stanza, stanzaType string, senderJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeAffiliations *cached.NodeAffiliations) bool {
	allowed := false

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
