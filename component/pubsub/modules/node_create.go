package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/satori/go.uuid"
	"strings"
	"github.com/ortuman/jackal-ff/component/pubsub/enums"
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

	instantNode := false
	instantNode = (nodeName == "")

	if instantNode {
		u1 := uuid.Must(uuid.NewV1())
		nodeName = strings.Replace(u1.String(), "-", "", -1)
	}

	// TODO
	// error if node exists
	//if getNodeConfig(toJID, nodeName) != nil {
	//	packet.(*xmpp.IQ).ConflictError()
	//}

	if toJID.IsFullWithUser() && toJID == packet.FromJID().ToBareJID() {
		return base.NewPubSubError(packet.(*xmpp.IQ).ForbiddenError())
	}

	nodeType := enums.Leaf
	//collection := ""


	nodeConfig := base.NewLeafNodeConfig(nodeName)
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
						nodeType = enums.Leaf
					} else {
						nodeType = enums.Collection
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
		return base.NewPubSubError(packet.(*xmpp.IQ).FeatureNotImplementedError())
	}

	if nodeType != enums.Leaf && nodeType != enums.Collection {
		return base.NewPubSubError(packet.(*xmpp.IQ).NotAllowedError())
	}

	// TODO
	// create node and store in DB
	// get Node Subscriptions
	// get Node Affiliations
	// subscribe if auto-subscribe
	// update Subscriptions and Affiliations

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
