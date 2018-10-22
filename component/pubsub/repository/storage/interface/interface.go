package _interface

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
)

type IPubSubDao interface {
	//GetNodeConfig(serviceJid string, nodeName string) (string, error)
	CreateNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeType int, collection string) (int64, error)
	GetNodeId(serviceJid jid.JID, nodeName string) (retNodeId int64)
	UpdateNodeConfig(jid jid.JID, nodeId int64, serializedData string, collectionId int64) (affectRows int64)
	GetNodeMeta(serviceJid jid.JID, nodeName string) (*model.NodeMeta, error)
}