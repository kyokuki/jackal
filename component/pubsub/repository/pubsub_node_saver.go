package repository

import (
	"github.com/ortuman/jackal/component/pubsub/repository/storage/interface"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
)

type nodeSaver struct{
	dao        _interface.IPubSubDao
}

func NewNodeSaver(paramDao _interface.IPubSubDao) *nodeSaver {
	nns := &nodeSaver{
		dao: paramDao,
	}
	return nns
}

func (nsr *nodeSaver) Save(node *cached.Node) error {
	var err error
	if node == nil {
		return nil
	}

	if node.IsDeleted() {
		return nil
	}

	if node.IsConfigNeedsWriting() {
		nsr.dao.UpdateNodeConfig(node.ServiceJid, node.GetNodeId(), node.NodeConfig.Form().Element().String(), 0)
	}

	if node.GetNodeAffiliations().AffiliationsNeedsWriting() {
		changedAffiliations := node.GetNodeAffiliations().GetChanged()
		for _, changedAff := range changedAffiliations {
			err = nsr.dao.SetNodeAffiliation(node.ServiceJid, node.GetNodeId(), node.Name, changedAff)
			if err != nil {
				return err
			}
		}
		node.GetNodeAffiliations().AffiliationsSaved()
	}

	if node.GetNodeSubscriptions().SubscriptionsNeedsWriting() {
		changedSubscriptions := node.GetNodeSubscriptions().GetChanged()
		for _, changedSub := range changedSubscriptions {
			err = nsr.dao.SetNodeSubscription(node.ServiceJid, node.GetNodeId(), node.Name, changedSub)
			if err != nil {
				return err
			}
		}
		node.GetNodeSubscriptions().SubscriptionsSaved()
	}

	return err
}
