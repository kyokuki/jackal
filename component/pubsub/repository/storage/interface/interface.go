package _interface

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp/jid"
)

type IPubSubDao interface {
	//GetNodeConfig(serviceJid string, nodeName string) (string, error)
	CreateNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeType int, collection string) (int64, error)
	GetNodeId(serviceJid jid.JID, nodeName string) (retNodeId int64)
}