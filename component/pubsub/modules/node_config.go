package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/module/xep0004"
)


type NodeConfigModule struct {}

func (s *NodeConfigModule)Name() string  {
	return "NodeConfigModule"
}

func (s *NodeConfigModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub#owner")
	eleCreate := &base.ElementCriteria{}
	eleCreate.SetName("configure")

	elePubsub.AddCriteria(eleCreate)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *NodeConfigModule)Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError  {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
	configure := pubsub.Elements().Child("configure")
	nodeName := configure.Attributes().Get("node")
	id := packet.ID()
	stanzaType := packet.Attributes().Get("type")


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


	// TODO
	// get affiation, and check if has the privileges : owner or admin
	// if not then error
	if false {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)
	}

	if "get" != stanzaType && "set" != stanzaType {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrBadRequest, nil)
	}

	if "get" == stanzaType {
		ps := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub#owner")
		rConfigure := xmpp.NewElementName("configure")
		rConfigure.SetAttribute("node", nodeName)

		elemForm := nodeConfig.Form()
		rConfigure.AppendElement(elemForm.Element())
		ps.AppendElement(rConfigure)

		resultStanza.AppendElement(ps)
	}

	if "set" == stanzaType {
		s.parseConf(nodeConfig, configure)
		// TODO collection node
		// update child nodes
		repository.Repository().UpdateNodeConfig(*toJID, nodeName, nodeConfig)
	}


	stm.SendElement(resultStanza)
	return nil
}

func (s *NodeConfigModule) parseConf(nodeConfig base.AbstractNodeConfig, configure xmpp.XElement) error {
	elementX := configure.Elements().ChildNamespace("x", "jabber:x:data")
	foo, err := xep0004.NewFormFromElement(elementX)
	if err != nil {
		return err
	}

	if elementX != nil && "submit" == elementX.Attributes().Get("type") {
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
