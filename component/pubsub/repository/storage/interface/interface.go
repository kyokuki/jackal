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
	GetNodeAffiliations(serviceJid jid.JID, nodeId int64) (*cached.NodeAffiliations, error)
	GetNodeSubscriptions(serviceJid jid.JID, nodeId int64) (*cached.NodeSubscriptions, error)
	DeleteNode(serviceJid jid.JID, nodeId int64) (error)
	GetChildNodes(serviceJid jid.JID, nodeName string) ([]string, error)

	GetUserSubscriptions(serviceJid jid.JID, userJid jid.JID) (map[string]*cached.NodeSubscriptions, error)
	GetUserAffiliations(serviceJid jid.JID, userJid jid.JID) (map[string]*cached.NodeAffiliations, error)

	GetItem(serviceJid jid.JID, nodeId int64, itemId string) (model.ItemMeta, error)
	QueryItems(nodeId int64, orderDate bool, orderAsc bool, limit int64) ([]model.ItemMeta, error)
	WriteItem(serviceJid jid.JID, nodeId int64, nodeName string, itemId string, publisherJid jid.JID, itemData string) (error)
	DeleteItem(serviceJid jid.JID, nodeId int64, itemId string) (error)
	GetItemIds(nodeId int64) ([]string, error)

	GetUserRoster(owner jid.JID) ([]model.UserRosterItem, error)
}
