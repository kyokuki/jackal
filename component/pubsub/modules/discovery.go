package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"fmt"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
	"strings"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type DiscoveryModule struct{
	DiscoInfo *xep0030.DiscoInfo
}

func (s *DiscoveryModule) Name() string {
	return "DiscoveryModule"
}

func (s *DiscoveryModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "get")

	eleQuery := &base.OrElementCriteria{}
	eleQuery1 := &base.ElementCriteria{}
	eleQuery1.SetName("query").AddAttr("xmlns", "http://jabber.org/protocol/disco#info")
	eleQuery2 := &base.ElementCriteria{}
	eleQuery2.SetName("query").AddAttr("xmlns", "http://jabber.org/protocol/disco#items")
	eleQuery.AddCriteria(eleQuery1)
	eleQuery.AddCriteria(eleQuery2)

	eleCrit.AddCriteria(eleQuery)
	return eleCrit
}

func (s *DiscoveryModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{}, nil
}

func (s *DiscoveryModule) Process(stanza xmpp.Stanza, stm stream.C2S) *base.PubSubError {
	var pubSubErr *base.PubSubError
	fmt.Println(s.Name() + " process ")

	query := stanza.Elements().Child("query")
	node := query.Attributes().Get("node")
	xmlns := query.Namespace()

	if node == "" {
		fmt.Println(xmlns)
		switch stanza := stanza.(type) {
		case *xmpp.IQ:
			if di := s.DiscoInfo; di != nil && di.MatchesIQ(stanza) {
				di.ProcessIQ(stanza, stm)
				return nil
			}
		}
		return nil
	}

	if xmlns == "http://jabber.org/protocol/disco#info" {
		pubSubErr = s.processDiscoInfo(stanza, stm)
	}

	if xmlns == "http://jabber.org/protocol/disco#items" {
		pubSubErr = s.processDiscoItems(stanza, stm)
	}

	return pubSubErr
}

func (s *DiscoveryModule) processDiscoInfo(stanza xmpp.XElement, stm stream.C2S) *base.PubSubError {
	stan := stanza.(xmpp.Stanza)
	fromJID := stan.FromJID()
	toJID := stan.ToJID()
	query := stan.Elements().ChildNamespace("query", "http://jabber.org/protocol/disco#info")
	nodeName := strings.Trim(query.Attributes().Get("node"), " ")

	resultIq := xmpp.NewElementName(stanza.Name())
	resultIq.SetTo(fromJID.String())
	resultIq.SetFrom(toJID.String())
	resultQuery := xmpp.NewElementNamespace("query", "http://jabber.org/protocol/disco#info")

	nodeMeta, _ := repository.Repository().GetNodeMeta(*toJID.ToBareJID(), nodeName)
	if nodeMeta == nil {
		return base.NewPubSubErrorStanza(stan, xmpp.ErrItemNotFound, nil)
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(stan, xmpp.ErrItemNotFound, nil)
	}
	clonedNodeConfig := nodeConfig.Clone()
	clonedForm := clonedNodeConfig.Form()
	clonedForm.AddField(xep0004.NewFieldHidden("FORM_TYPE", "http://jabber.org/protocol/pubsub#meta-data"))

	var owners []string
	var publishers [] string

	affiliations := repository.Repository().GetNodeAffiliations(*toJID.ToBareJID(), nodeName)
	for _, userAff := range affiliations.GetAffiliations() {
		switch userAff.GetAffiliation() {
		case enums.AffiliationOwner:
			owners = append(owners, userAff.GetJid().String())
		case enums.AffiliationPublisher:
			publishers = append(publishers, userAff.GetJid().String())
		default:
			// do nothing
		}
	}
	clonedForm.AddField(xep0004.NewFieldJidMulti("pubsub#owner", owners, "Node owners"))
	clonedForm.AddField(xep0004.NewFieldJidMulti("pubsub#publisher", publishers, "Publishers to this node"))
	clonedForm.AddField(xep0004.NewFieldJidSingle("pubsub#creator", nodeMeta.Creator, "Node creator")) // TODO
	clonedForm.AddField(xep0004.NewFieldJidSingle("pubsub#creation_date", nodeMeta.CreateDate.Format("2006-01-02T15:04:05Z"), "Creation date")) // TODO

	elemIdentity := xmpp.NewElementName("identity")
	elemIdentity.SetAttribute("category", "pubsub")
	elemIdentity.SetAttribute("type", clonedNodeConfig.GetNodeType().String())
	elemFeature := xmpp.NewElementName("feature ")
	elemFeature.SetAttribute("var", "http://jabber.org/protocol/pubsub")

	resultQuery.AppendElement(elemIdentity)
	resultQuery.AppendElement(elemFeature)
	resultQuery.AppendElement(clonedForm.Element())

	resultIq.AppendElement(resultQuery)
	stm.SendElement(resultIq)
	return nil
}

func (s *DiscoveryModule) processDiscoItems(stanza xmpp.XElement, stm stream.C2S) *base.PubSubError {
	stan := stanza.(xmpp.Stanza)
	fromJID := stan.FromJID()
	toJID := stan.ToJID()

	resultIq := xmpp.NewElementName(stanza.Name())
	resultIq.SetTo(fromJID.String())
	resultIq.SetFrom(toJID.String())
	resultQuery := xmpp.NewElementNamespace("query", "http://jabber.org/protocol/disco#items")

	resultIq.AppendElement(resultQuery)
	stm.SendElement(resultIq)
	return nil
}