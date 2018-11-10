package sql

import (
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"github.com/ortuman/jackal/component/pubsub/enums"
)


func (s *Storage) GetUserSubscriptions(serviceJid jid.JID, userJid jid.JID) (map[string]*cached.NodeSubscriptions, error) {
	var err error
	rows, err := s.db.Query(`
		select n.name, ps.subscription, ps.subscription_id from pubsub_nodes n
		inner join pubsub_service_jids sj on sj.service_id = n.service_id
		inner join pubsub_subscriptions ps on ps.node_id = n.node_id
		inner join pubsub_jids pj on pj.jid_id = ps.jid_id
		where pj.jid_sha1 = ? and sj.service_jid_sha1 = ?
			and pj.jid = ? and sj.service_jid = ?`,
		s.Sha1(userJid.ToBareJID().String()),
		s.Sha1(serviceJid.ToBareJID().String()),
		userJid.ToBareJID().String(),
		serviceJid.ToBareJID().String())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var resultMap map[string]*cached.NodeSubscriptions
	resultMap = make(map[string]*cached.NodeSubscriptions)
	for rows.Next() {
		var (
			scanNodeName   string
			scanSubscription   string
			scanSubid string
		)
		err = rows.Scan(&scanNodeName, &scanSubscription, &scanSubid)
		if err != nil {
			return nil, err
		}

		if _, ok := resultMap[scanNodeName]; !ok {
			resultMap[scanNodeName] = cached.NewNodeSubscriptions()
		}
		nodeSubscription, _ := resultMap[scanNodeName]
		nodeSubscription.AddSubscription(userJid, enums.NewSubscriptionValue(scanSubscription), scanSubid)
	}
	return resultMap, nil
}