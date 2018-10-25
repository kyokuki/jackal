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

	nodeId			int64
	deleted           bool
	nodeAffiliations  *NodeAffiliations
	nodeSubscriptions *NodeSubscriptions
}

func NewNode(nodeId int64, serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, date time.Time) Node {
	node := Node{
		nodeId:     nodeId,
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

func (nd *Node) SetNodeAffiliations(newNodeAffiliations *NodeAffiliations) {
	nd.nodeAffiliations = newNodeAffiliations
}

func (nd *Node) GetNodeSubscriptions() *NodeSubscriptions {
	return nd.nodeSubscriptions
}

func (nd *Node) SetNodeSubscriptions(newNodeSubscriptions *NodeSubscriptions) {
	nd.nodeSubscriptions = newNodeSubscriptions
}

func (nd *Node) SetDeleted()  {
	nd.deleted = true
}

func (nd *Node) IsDeleted() bool {
	return nd.deleted
}

func (nd *Node) GetNodeId() int64 {
	return nd.nodeId
}
