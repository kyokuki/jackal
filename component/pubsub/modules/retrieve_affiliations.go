package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"strings"
	"log"
)

type RetrieveAffiliationsModule struct{}

func (s *RetrieveAffiliationsModule) Name() string {
	return "RetrieveAffiliationsModule"
}

func (s *RetrieveAffiliationsModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleAffiliations := &base.ElementCriteria{}
	eleAffiliations.SetName("affiliations")

	elePubsub.AddCriteria(eleAffiliations)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *RetrieveAffiliationsModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#retrieve-affiliations",
		"http://jabber.org/protocol/pubsub#publisher-affiliation",
		"http://jabber.org/protocol/pubsub#outcast-affiliation",
		"http://jabber.org/protocol/pubsub#member-affiliation",
	}, nil
}

func (s *RetrieveAffiliationsModule) Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	affiliations := pubsub.Elements().Child("affiliations")
	nodeName := strings.Trim(affiliations.Attributes().Get("node"), " ")
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	pubSubResult := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	subAffResult := xmpp.NewElementName("affiliations")

	if nodeName == "" {
		nodeUserAffiliationMap, err := repository.Repository().GetUserAffiliations(*toJID.ToBareJID(), *fromJID.ToBareJID())
		if err != nil {
			log.Print(err.Error())
		}
		for tmpNodeName, item := range nodeUserAffiliationMap {
			affiliations := item.GetAffiliations()
			for _, item := range affiliations {
				elemAff := xmpp.NewElementName("affiliation")
				elemAff.SetAttribute("node", tmpNodeName)
				elemAff.SetAttribute("subscription", item.GetAffiliation().String())
				subAffResult.AppendElement(elemAff)
			}
		}
	} else {
		nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
		if nodeConfig == nil {
			return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
		}

		subAffResult.SetAttribute("node", nodeName)
		nodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID.ToBareJID(), nodeName)
		affs := nodeAffiliations.GetAffiliations()
		for _, item := range affs {
			if item.GetJid().ToBareJID().String() != fromJID.ToBareJID().String() {
				continue
			}
			elemAff := xmpp.NewElementName("affiliation")
			elemAff.SetAttribute("node", nodeName)
			elemAff.SetAttribute("subscription", item.GetAffiliation().String())

			subAffResult.AppendElement(elemAff)
		}
	}

	pubSubResult.AppendElement(subAffResult)
	resultStanza.AppendElement(pubSubResult)
	stm.SendElement(resultStanza)
	return nil
}
