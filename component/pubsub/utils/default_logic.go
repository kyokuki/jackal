package utils

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
)

type privatePubSubLogic struct {

}

var DefaultPubSubLogic privatePubSubLogic

func (psl *privatePubSubLogic) HasSenderSubscription(jid jid.JID, nodeAffiations *cached.NodeAffiliations, nodeSubscriptions *cached.NodeSubscriptions) bool {
	return false
}

func (psl *privatePubSubLogic) IsSenderInRosterGroup(jid jid.JID, nodeConfig base.AbstractNodeConfig, nodeAffiations *cached.NodeAffiliations, nodeSubscriptions *cached.NodeSubscriptions) bool {
	return false
}