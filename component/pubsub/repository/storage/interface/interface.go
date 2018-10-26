package _interface

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
)

type IPubSubDao interface {
	//GetNodeConfig(serviceJid string, nodeName string) (string, error)
	CreateNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeType int, collection string) (nodeId int64, err error)
	GetNodeId(serviceJid jid.JID, nodeName string) (nodeId int64)
	UpdateNodeConfig(jid jid.JID, nodeId int64, serializedData string, collectionId int64) (affectRows int64, err error)
	GetNodeMeta(serviceJid jid.JID, nodeName string) (*model.NodeMeta, error)
	SetNodeAffiliation(serviceJid jid.JID, nodeId int64, nodeName string, affiliation stateless.UsersAffiliation) (error)
	SetNodeSubscription(serviceJid jid.JID, nodeId int64, nodeName string, subscription stateless.UsersSubscription) (error)
	GetNodeAffiliations(serviceJid jid.JID, nodeId int64) (*cached.NodeAffiliations ,error)
	GetNodeSubscriptions(serviceJid jid.JID, nodeId int64) (*cached.NodeSubscriptions ,error)
	DeleteNode(serviceJid jid.JID, nodeId int64) (error)
}