package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
)


type NodeDeleteModule struct {}

func (s *NodeDeleteModule)Name() string  {
	return "NodeDeleteModule"
}

func (s *NodeDeleteModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq")
	eleCrit.AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	eleDelete := &base.ElementCriteria{}
	eleDelete.SetName("delete")

	elePubsub.AddCriteria(eleDelete)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *NodeDeleteModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#delete-nodes",
	}, nil
}

func (s *NodeDeleteModule)Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError  {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	delete := pubsub.Elements().Child("delete")
	nodeName := delete.Attributes().Get("node")
	id := packet.ID()

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")
	resultStanza.SetAttribute("id", id)

	if nodeName == "" {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrNotAllowed,
			[]xmpp.XElement{})
	}

	nodeConfig := repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig == nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrItemNotFound, nil)
	}


	tmpNodeAffiliations := repository.Repository().GetNodeAffiliations(*toJID, nodeName)
	// TODO when jid is admin, do not check privileges
	if !false {
		if tmpNodeAffiliations != nil {
			senderAffiliation := tmpNodeAffiliations.GetSubscriberAffiliation(*fromJID)
			if !senderAffiliation.GetAffiliation().IsDeleteNode() {
				return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)
			}
		}
	}

	// TODO [pubsub#notify_config]
	//_, notify_config := nodeConfig.Form().Field("pubsub#notify_config")
	//if len(notify_config.Values) > 0 && notify_config.Values[0] == "1" {
	//
	//}

	repository.Repository().DeleteNode(*toJID, nodeName)

	stm.SendElement(resultStanza)
	return nil
}

func (s *NodeDeleteModule) parseConf(nodeConfig base.AbstractNodeConfig, configure xmpp.XElement) error {
	elementX := configure.Elements().ChildNamespace("x", "jabber:x:data")

	if elementX != nil && "submit" == elementX.Attributes().Get("type") {
		foo, err := xep0004.NewFormFromElement(elementX)
		if err != nil {
			return err
		}

		for _, itemField := range foo.Fields {
			variable := itemField.Var


			if variable == "pubsub#send_last_published_item" {
				value := ""
				if len(itemField.Values) > 0 {
					value = itemField.Values[0]
				}

				// TODO
				// "Requested on_sub_and_presence mode for sending last published item is disabled."
				if value == "xxxxx" {

				}
			}

			nodeConfig.Form().AddField(itemField)
		}
	}

	return nil
}
