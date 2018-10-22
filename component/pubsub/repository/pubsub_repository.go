package repository

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/base"
	"time"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"fmt"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/interface"
	"github.com/ortuman/jackal/component/pubsub/repository/storage"
)

type pubSubRepository struct {
	nodes map[cached.NodeKey]*cached.Node
	dao _interface.IPubSubDao
	nodesAdded int64
}

var instancePubSubRepository pubSubRepository

func init()  {
	instancePubSubRepository.nodes = make(map[cached.NodeKey]*cached.Node)
}

func Init(mysql string) {
	storage.InitStorage(mysql)
	instancePubSubRepository.dao = storage.Instance()
}

func Repository() *pubSubRepository {
	return &instancePubSubRepository
}

func (ps *pubSubRepository) CreateNode(
	bareJid jid.JID,
	nodeName string,
	ownerJid jid.JID,
	nodeConfig base.AbstractNodeConfig,
	nodeType string,
	collection string) error {

		// TODO
		// 1. check parent collection
		// err : Parent collection does not exists yet!



		// 2. create node in DB
		// err : Creating node failed!
		intNodeType := 2
		if nodeType != "collection" {
			intNodeType = 1
		}
		retNodeId, err := ps.dao.CreateNode(bareJid, nodeName, ownerJid, nodeConfig, intNodeType, collection)
		if err != nil {
			return err
		}

		retNodeId2 := ps.dao.GetNodeId(bareJid, nodeName)
		if retNodeId2 < 0 {
			return fmt.Errorf("Creating node failed!")
		}

		node := cached.NewNode(retNodeId, bareJid, nodeName, ownerJid, nodeConfig, time.Now())
		nodeKey := cached.NewNodeKey(bareJid.ToBareJID().String(), nodeName)
		ps.nodes[nodeKey] = &node

		// TODO
		// get NodeAffiliations and NodeSubscriptions, and store them in the node which is created above

		ps.nodesAdded += 1
		return nil
}

func (ps *pubSubRepository) GetNodeConfig(serviceJid jid.JID,nodeName string) base.AbstractNodeConfig {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return nil
	}
	if node == nil {
		return nil
	}
	return node.NodeConfig
}

func (ps *pubSubRepository) DeleteNode(serviceJid jid.JID,nodeName string) error {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}


	// TODO
	// delete Node Info from DB ...

	nodeKey := cached.NewNodeKey(serviceJid.ToBareJID().String(), nodeName)
	delete(ps.nodes, nodeKey)
	node.SetDeleted()
	return nil
}

func (ps *pubSubRepository) getNode(serviceJid jid.JID, nodeName string) (*cached.Node, error) {
	nodeKey := cached.NewNodeKey(serviceJid.ToBareJID().String(), nodeName)
	node, ok := ps.nodes[nodeKey]
	if ok {
		return node, nil
	}

	// TODO
	// get node info from DB, construct a node-struct and return it

	return nil, nil
}

func (ps *pubSubRepository) UpdateNodeConfig(serviceJid jid.JID, nodeName string, nodeConfig base.AbstractNodeConfig) (error) {
	nodeKey := cached.NewNodeKey(serviceJid.ToBareJID().String(), nodeName)
	node, ok := ps.nodes[nodeKey]
	if !ok {
		node.NodeConfig.Form().CopyValuesFromDataForm(nodeConfig.Form())

		// TODO  save node config in DB
	}
	return nil
}

func (ps *pubSubRepository) GetNodeAffiliations(serviceJid jid.JID, nodeName string) *cached.NodeAffiliations {
	node, _ := ps.getNode(serviceJid, nodeName)
	if node != nil {
		return node.GetNodeAffiliations()
	}
	return nil
}


func (ps *pubSubRepository) GetNodeSubscriptions(serviceJid jid.JID, nodeName string) *cached.NodeSubscriptions {
	node, _ := ps.getNode(serviceJid, nodeName)
	if node != nil {
		return node.GetNodeSubscriptions()
	}
	return nil
}


func (ps *pubSubRepository) UpdateNodeAffiliations(serviceJid jid.JID, nodeName string, nodeAffiliations *cached.NodeAffiliations) (error) {
	node, _ := ps.getNode(serviceJid, nodeName)
	if node != nil {
		// pointer is not match
		if node.GetNodeAffiliations() != nodeAffiliations {
			return fmt.Errorf("INCORRECT")
		}

		// TODO write to DB
	}

	return nil
}

func (ps *pubSubRepository) UpdateNodeSubscriptions(serviceJid jid.JID, nodeName string, nodeSubscriptions *cached.NodeSubscriptions) (error) {
	node, _ := ps.getNode(serviceJid, nodeName)
	if node != nil {
		// pointer is not match
		if node.GetNodeSubscriptions() != nodeSubscriptions {
			return fmt.Errorf("INCORRECT")
		}

		// TODO write to DB
	}

	return nil
}



