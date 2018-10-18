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
}

type abstractNodeConfig struct {
	isInit   bool
	form     xep0004.DataForm
	nodeName string
}

func (af *abstractNodeConfig) Form() *xep0004.DataForm {
	if !af.isInit {
		af.init("default")
	}
	return &af.form
}

func (af *abstractNodeConfig) init(nodeName string) {
	af.nodeName = nodeName
	af.initForm()
	af.isInit = true
}

func (af *abstractNodeConfig) initForm() {
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
	af.form.AddField(xep0004.NewFieldBool(PUBSUB+"presence_based_delivery", true,
		"Whether to deliver notifications to available users only"))
}
