package base

import (
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

const (
	PUBSUB = "pubsub#"
)

type AbstractNodeConfig interface {
	Form() *xep0004.DataForm
	IsNotifyConfig() bool
	GetNodeAccessModel() enums.AccessModelType
	GetNodeType() enums.NodeType
	GetPublisherModel() enums.PublisherModelType
	IsDeliverPresenceBased() bool

	Clone() AbstractNodeConfig
	GetRosterGroupsAllowed() []string
}

type NodeConfigType struct {
	isInit   bool
	form     xep0004.DataForm
	nodeName string
}

func (af *NodeConfigType) Form() *xep0004.DataForm {
	if !af.isInit {
		af.init("default")
	}
	return &af.form
}

func (af *NodeConfigType) SetForm(form *xep0004.DataForm) {
	af.form = *form
}

func (af *NodeConfigType) init(nodeName string) {
	af.nodeName = nodeName
	af.initForm()
	af.isInit = true
}

func (af *NodeConfigType) initForm() {
	af.form.Type = xep0004.Form

	newField := xep0004.Field{}
	af.form.AddField(xep0004.NewFieldHidden("FORM_TYPE", "http://jabber.org/protocol/pubsub#node_config"))

	newField, _ = xep0004.NewFieldListSingle(PUBSUB+"node_type", "leaf", "", []string{}, []string{enums.Leaf.String(), enums.Collection.String()})
	af.form.AddField(newField)

	af.form.AddField(xep0004.NewFieldTextSingle(PUBSUB+"title", "", "A friendly name for the node"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"deliver_payloads", true,
		"A friendly name for the nodeWhether to deliver payloads with event notifications"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"notify_config", false,
		"Notify subscribers when the node configuration changes"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"notify_delete", false,
		"Notify subscribers when the node is deleted"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"notify_retract", false,
		"Notify subscribers when items are removed from the node"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"persist_items", true, "Persist items to storage"))
	af.form.AddField(xep0004.NewFieldTextSingle(PUBSUB+"max_items", "10", "Max # of items to persist"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"subscribe", true, "Whether to allow subscriptions"))
	af.form.AddField(xep0004.NewFieldTextSingle(PUBSUB+"collection", "",
		"The collection with which a node is affiliated"))

	newField, _ = xep0004.NewFieldListSingle(PUBSUB+"access_model", enums.AccessModelOpen.String(),
		"Specify the subscriber model",
		[]string{},
		[]string{
			enums.AccessModelAuthorize.String(),
			enums.AccessModelOpen.String(),
			enums.AccessModelPresence.String(),
			enums.AccessModelRoster.String(),
			enums.AccessModelWhitelist.String(),
		})
	af.form.AddField(newField)

	newField, _ = xep0004.NewFieldListSingle(PUBSUB+"publish_model", enums.PublisherModelPublishers.String(),
		"Specify the publisher model",
		[]string{},
		[]string{
			enums.PublisherModelPublishers.String(),
			enums.PublisherModelSubscribers.String(),
			enums.PublisherModelOpen.String(),
		})
	af.form.AddField(newField)

	newField, _ = xep0004.NewFieldListSingle(PUBSUB+"send_last_published_item", enums.SendLastPublishedItem_on_sub.String(),
		"When to send the last published item",
		[]string{},
		[]string{
			enums.SendLastPublishedItem_never.String(),
			enums.SendLastPublishedItem_on_sub.String(),
			enums.SendLastPublishedItem_on_sub_and_presence.String(),
		})
	af.form.AddField(newField)

	af.form.AddField(xep0004.NewFieldTextMulti(PUBSUB+"domains", []string{},
		"The domains allowed to access this node (blank for any)"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"presence_based_delivery", false,
		"Whether to deliver notifications to available users only"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"presence_expired", false,
		"Whether to subscription expired when subscriber going offline."))
	af.form.AddField(xep0004.NewFieldTextMulti(PUBSUB+"embedded_body_xslt", []string{},
		"The XSL transformation which can be applied to payloads in order to generate an appropriate message body element."))
	af.form.AddField(xep0004.NewFieldTextSingle(PUBSUB+"body_xslt", "",
		"The URL of an XSL transformation which can be applied to payloads in order to generate an appropriate message body element."))
	af.form.AddField(xep0004.NewFieldTextMulti(PUBSUB+"roster_groups_allowed", []string{},
		"Roster groups allowed to subscribe"))
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"notify_sub_aff_state", false,
		"Notify subscribers when owner change their subscription or affiliation state"))
}


func (af *NodeConfigType) IsNotifyConfig() bool {
	_, notifyConfig := af.Form().Field("pubsub#notify_config")
	if len(notifyConfig.Values) > 0 && notifyConfig.Values[0] == "1" {
		return true
	}
	return false
}

func (af *NodeConfigType) GetNodeAccessModel() enums.AccessModelType {
	_, accessModel := af.Form().Field("pubsub#access_model")
	if len(accessModel.Values) > 0 {
		return enums.AccessModelType(accessModel.Values[0])
	}
	return enums.AccessModelType("")
}

func (af *NodeConfigType) GetNodeType() enums.NodeType {
	_, nodeType := af.Form().Field("pubsub#node_type")
	if len(nodeType.Values) > 0 {
		return enums.NewNodeType(nodeType.Values[0])
	}
	return enums.NewNodeType("")
}

func (af *NodeConfigType) GetPublisherModel() enums.PublisherModelType {
	_, nodeType := af.Form().Field("pubsub#publish_model")
	if len(nodeType.Values) > 0 {
		return enums.NewPublisherModelType(nodeType.Values[0])
	}
	return enums.NewPublisherModelType("")
}

func (af *NodeConfigType) IsDeliverPresenceBased() bool {
	_, delivery := af.Form().Field("pubsub#presence_based_delivery")
	if len(delivery.Values) > 0 {
		if delivery.Values[0] =="true" || delivery.Values[0] == "1" {
			return true
		}
	}
	return false
}

func (af *NodeConfigType) Clone() AbstractNodeConfig {
	var ins AbstractNodeConfig
	if af.GetNodeType() == enums.Leaf {
		ins = NewLeafNodeConfig(af.nodeName)
	} else if af.GetNodeType() == enums.Collection {
		ins = NewCollectionNodeConfig(af.nodeName)
	} else {
		ins = NewLeafNodeConfig(af.nodeName)
	}
	ins.Form().CopyValuesFromDataForm(af.Form())
	return ins
}

func (af *NodeConfigType) GetRosterGroupsAllowed() []string {
	_, rosterGroupsAllowed := af.Form().Field("pubsub#roster_groups_allowed")
	return rosterGroupsAllowed.Values
}