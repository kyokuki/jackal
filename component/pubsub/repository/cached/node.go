package cached

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"time"
	"github.com/ortuman/jackal/xmpp/jid"
)

type NodeMeta interface {
	GetNodeConfig() base.AbstractNodeConfig
	GetCreator() string
	GetCreationTime() time.Time
}

type Node struct {
	Date              time.Time
	Creator           jid.JID
	Name              string
	ServiceJid        jid.JID
	NodeConfig        base.AbstractNodeConfig

	nodeAffiliations  *NodeAffiliations
	nodeSubscriptions *NodeSubscriptions
}

func NewNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, date time.Time) Node {
	node := Node{
		Date:       date,
		Name:       nodeName,
		ServiceJid: serviceJid,
		Creator:    ownerJid,
		NodeConfig: nodeConfig,
	}
	node.nodeAffiliations = NewNodeAffiliations()
	node.nodeSubscriptions = NewNodeSubscriptions()
	return node
}

type NodeKey struct {
	ServiceJid string
	Node       string
}

func NewNodeKey(bareJid string, nodeName string) NodeKey {
	return NodeKey{bareJid, nodeName}
}

func (nd *Node) GetNodeAffiliations() *NodeAffiliations {
	return nd.nodeAffiliations
}

func (nd *Node) GetNodeSubscriptions() *NodeSubscriptions {
	return nd.nodeSubscriptions
}
