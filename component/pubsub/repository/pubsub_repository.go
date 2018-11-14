package repository

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/base"
	"time"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"fmt"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/interface"
	"github.com/ortuman/jackal/component/pubsub/repository/storage"
	"strings"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
)

type pubSubRepository struct {
	nodes      map[cached.NodeKey]*cached.Node
	dao        _interface.IPubSubDao
	nodesAdded int64
	nodeSaver  *nodeSaver
}

var instancePubSubRepository pubSubRepository

func Init(mysql string) {
	instancePubSubRepository.nodes = make(map[cached.NodeKey]*cached.Node)
	storage.InitStorage(mysql)
	instancePubSubRepository.dao = storage.Instance()
	instancePubSubRepository.nodeSaver = NewNodeSaver(instancePubSubRepository.dao)
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

func (ps *pubSubRepository) GetNodeConfig(serviceJid jid.JID, nodeName string) base.AbstractNodeConfig {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return nil
	}
	if node == nil {
		return nil
	}
	return node.NodeConfig
}

func (ps *pubSubRepository) DeleteNode(serviceJid jid.JID, nodeName string) error {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}

	ps.dao.DeleteNode(serviceJid, node.GetNodeId())

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

	nodeMeta, err := ps.dao.GetNodeMeta(serviceJid, nodeName)
	if err != nil {
		return nil, err
	}

	xmppParser := xmpp.NewParser(strings.NewReader(nodeMeta.NodeConfig), xmpp.DefaultMode, 0)
	nodeConfigElement, err := xmppParser.ParseElement()
	if err != nil {
		return nil, err
	}
	nodeConfigForm, err := xep0004.NewFormFromElement(nodeConfigElement)
	if err != nil {
		return nil, err
	}

	nodeConfig := base.NewLeafNodeConfig(nodeName)
	nodeConfig.SetForm(nodeConfigForm)

	nodeAffiliations, err := ps.dao.GetNodeAffiliations(serviceJid, nodeMeta.NodeId)
	nodeSubscriptions, err := ps.dao.GetNodeSubscriptions(serviceJid, nodeMeta.NodeId)

	creatorJid, _ := jid.NewWithString(nodeMeta.Creator, true)
	newNode := cached.NewNode(nodeMeta.NodeId, serviceJid, nodeName, *creatorJid, nodeConfig, nodeMeta.CreateDate)

	newNode.SetNodeAffiliations(nodeAffiliations)
	newNode.SetNodeSubscriptions(nodeSubscriptions)

	newNodeKey := cached.NewNodeKey(serviceJid.ToBareJID().String(), nodeName)
	ps.nodes[newNodeKey] = &newNode

	return &newNode, nil
}

func (ps *pubSubRepository) UpdateNodeConfig(serviceJid jid.JID, nodeName string, nodeConfig base.AbstractNodeConfig) (error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err == nil && node != nil {
		node.ConfigCopyFrom(nodeConfig)
		ps.nodeSaver.Save(node)
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

		ps.nodeSaver.Save(node)
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

		ps.nodeSaver.Save(node)
	}

	return nil
}

func (ps *pubSubRepository) GetUserSubscriptions(serviceJid jid.JID, userJid jid.JID) (map[string]*cached.NodeSubscriptions, error) {
	return ps.dao.GetUserSubscriptions(serviceJid, userJid)
}

func (ps *pubSubRepository) GetUserAffiliations(serviceJid jid.JID, userJid jid.JID) (map[string]*cached.NodeAffiliations, error) {
	return ps.dao.GetUserAffiliations(serviceJid, userJid)
}

func (ps *pubSubRepository) GetItem(serviceJid jid.JID, nodeName string, itemId string) (model.ItemMeta, error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return model.ItemMeta{}, err
	}
	return ps.dao.GetItem(serviceJid, node.GetNodeId(), itemId)
}

func (ps *pubSubRepository) QueryItems(serviceJid jid.JID, nodeName string, maxItems int64) ([]model.ItemMeta, error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return nil, err
	}
	if maxItems == 0 {
		maxItems = 10
	}
	return ps.dao.QueryItems(node.GetNodeId(), true, true, maxItems)
}

func (ps *pubSubRepository) WriteItem(serviceJid jid.JID, nodeName string, itemId string, publisherJid jid.JID, itemElem xmpp.XElement) (error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return err
	}
	return ps.dao.WriteItem(serviceJid, node.GetNodeId(), nodeName, itemId, publisherJid, itemElem.String())
}

func (ps *pubSubRepository) DeleteItem(serviceJid jid.JID, nodeName string, itemId string) (error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return err
	}
	return ps.dao.DeleteItem(serviceJid, node.GetNodeId(), itemId)
}

func (ps *pubSubRepository) GetItemIds(serviceJid jid.JID, nodeName string) ([]string, error) {
	node, err := ps.getNode(serviceJid, nodeName)
	if err != nil {
		return nil, err
	}
	return ps.dao.GetItemIds(node.GetNodeId())
}
