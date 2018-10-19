package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/satori/go.uuid"
	"strings"
	"github.com/ortuman/jackal-ff/component/pubsub/enums"
	"github.com/ortuman/jackal/component/pubsub/repository"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
)


type NodeCreateModule struct {}

func (s *NodeCreateModule)Name() string  {
	return "NodeCreateModule"
}

func (s *NodeCreateModule) ModuleCriteria() *base.ElementCriteria {
	eleCrit := &base.ElementCriteria{}
	eleCrit.SetName("iq").AddAttr("type", "set")
	elePubsub := &base.ElementCriteria{}
	elePubsub.SetName("pubsub").AddAttr("xmlns", "http://jabber.org/protocol/pubsub")
	eleCreate := &base.ElementCriteria{}
	eleCreate.SetName("create")

	elePubsub.AddCriteria(eleCreate)
	eleCrit.AddCriteria(elePubsub)
	return eleCrit
}

func (s *NodeCreateModule) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{
		"http://jabber.org/protocol/pubsub#create-nodes",
		"http://jabber.org/protocol/pubsub#instant-nodes",
		"http://jabber.org/protocol/pubsub#create-and-configure",
	}, nil
}


func (s *NodeCreateModule)Process(packet xmpp.Stanza, stm stream.C2S) *base.PubSubError  {
	fromJID := packet.FromJID()
	toJID := packet.ToJID().ToBareJID()
	pubsub := packet.Elements().ChildNamespace("pubsub", "http://jabber.org/protocol/pubsub")
	create := pubsub.Elements().Child("create")
	configure := pubsub.Elements().Child("configure")
	nodeName := create.Attributes().Get("node")

	resultStanza := xmpp.NewElementName(packet.Name())
	resultStanza.SetTo(fromJID.String())
	resultStanza.SetFrom(toJID.String())
	resultStanza.SetAttribute("type", "result")

	instantNode := false
	instantNode = (nodeName == "")

	if instantNode {
		u1 := uuid.Must(uuid.NewV1())
		nodeName = strings.Replace(u1.String(), "-", "", -1)
	}

	// error if node exists
	var nodeConfig base.AbstractNodeConfig
	nodeConfig = repository.Repository().GetNodeConfig(*toJID, nodeName)
	if nodeConfig != nil {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrConflict, nil)
	}

	if toJID.IsFullWithUser() && toJID == packet.FromJID().ToBareJID() {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrForbidden, nil)

	}

	nodeType := enums.Leaf
	collection := ""


	nodeConfig = base.NewLeafNodeConfig(nodeName)
	if configure != nil {
		elementX := configure.Elements().ChildNamespace("x", "jabber:x:data")

		if elementX != nil && "submit" == elementX.Attributes().Get("type") {
			for _, elementField := range elementX.Elements().All() {
				if elementField.Name() != "field" {
					continue
				}

				variable := elementField.Attributes().Get("var")
				val := ""
				elementValue := elementField.Elements().Child("value")
				if elementValue != nil {
					val = elementValue.Text()
				}

				if "pubsub#node_type" == variable {
					if val == enums.Collection.String() {
						nodeType = enums.Collection
					} else {
						nodeType = enums.Leaf
					}
				}

				// TODO implement [pubsub#send_last_published_item]
				// ...

				nodeConfig.Form().SetField(variable, val)
			}
		}
	}

	// currently not suport Collection Feature
	if nodeType == enums.Collection {
		unsupported1 := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		unsupported1.SetAttribute("feature", "collections")
		unsupported2 := xmpp.NewElementNamespace("unsupported", "http://jabber.org/protocol/pubsub#errors")
		unsupported2.SetAttribute("feature", "multi-collection")
		return base.NewPubSubErrorStanza(packet, xmpp.ErrFeatureNotImplemented,
			[]xmpp.XElement{
				unsupported1,
				unsupported2,
			})
	}

	if nodeType != enums.Leaf && nodeType != enums.Collection {
		return base.NewPubSubErrorStanza(packet, xmpp.ErrNotAllowed, nil)
	}

	// TODO
	// get Node Subscriptions
	// get Node Affiliations
	// subscribe if auto-subscribe
	// update Subscriptions and Affiliations
	repository.Repository().CreateNode(*toJID, nodeName, *fromJID.ToBareJID(), nodeConfig, nodeType.String(), collection)


	if instantNode {
		ps := xmpp.NewElementNamespace("pubsub", "http://jabber.org/protocol/pubsub")
		cr := xmpp.NewElementName("create")
		cr.SetAttribute("node", nodeName)
		ps.AppendElement(cr)

		resultStanza.AppendElement(ps)
	}

	stm.SendElement(resultStanza)
	return nil
}
