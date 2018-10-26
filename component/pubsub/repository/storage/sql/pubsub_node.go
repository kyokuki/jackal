package sql

import (
	sq "github.com/Masterminds/squirrel"
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
	"time"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
)

func (s *Storage) CreateNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeType int, collection string) (int64, error) {
	var err error
	var retNodeId int64 = -1

	serializedNodeConfig := ""
	if nodeConfig != nil {
		serializedNodeConfig = nodeConfig.Form().Element().String()
	}

	serviceId := s.PubSubEnsureServiceJid(serviceJid.String())
	jidId := s.PubSubEnsureJid(ownerJid.String())

	qSelect := sq.Select("node_id").
		From("pubsub_nodes").
		Where(sq.Eq{
		"name":       nodeName,
		"service_id": serviceId,
	})

	err = qSelect.RunWith(s.db).QueryRow().Scan(&retNodeId)
	if err == sql.ErrNoRows {
		columns := []string{"service_id", "name", "name_sha1", "type", "creator_id", "creation_date", "configuration", "collection_id"}
		values := []interface{}{serviceId, nodeName, s.Sha1(nodeName), nodeType, jidId, time.Now(), serializedNodeConfig, collection}

		qInsert := sq.Insert("pubsub_nodes").
			Columns(columns...).
			Values(values...)
		sqlRet, err := qInsert.RunWith(s.db).Exec()
		if err == nil {
			retNodeId, err = sqlRet.LastInsertId()
		}
	}

	return retNodeId, err
}

func (s *Storage) GetNodeId(serviceJid jid.JID, nodeName string) (retNodeId int64) {
	retNodeId = -1
	err := s.db.QueryRow(`
				select n.node_id from pubsub_nodes n 
				inner join pubsub_service_jids sj on n.service_id = sj.service_id
				where sj.service_jid = ? and n.name = ?`, serviceJid.String(), nodeName).Scan(&retNodeId)
	if err != nil {
		retNodeId = -1
	}
	return
}

func (s *Storage) PubSubEnsureServiceJid(serviceJid string) (retServiceId int64) {
	retServiceId = -1
	qSelect := sq.Select("service_id").
		From("pubsub_service_jids").
		Where(sq.Eq{
		"service_jid":      serviceJid,
		"service_jid_sha1": s.Sha1(serviceJid),
	})

	var qServiceId int64
	err := qSelect.RunWith(s.db).QueryRow().Scan(&qServiceId)
	if err == sql.ErrNoRows {
		qInsert := sq.Insert("pubsub_service_jids").
			Columns([]string{"service_jid", "service_jid_sha1"}...).
			Values([]interface{}{serviceJid, s.Sha1(serviceJid)}...)

		sqlRet, err := qInsert.RunWith(s.db).Exec()
		if err == nil {
			retServiceId, _ = sqlRet.LastInsertId()
			return
		}
	}
	retServiceId = qServiceId
	return
}

func (s *Storage) PubSubEnsureJid(jid string) (retJidId int64) {
	retJidId = -1
	qSelect := sq.Select("jid_id").
		From("pubsub_jids").
		Where(sq.Eq{
		"jid":      jid,
		"jid_sha1": s.Sha1(jid),
	})

	var qJidId int64
	err := qSelect.RunWith(s.db).QueryRow().Scan(&qJidId)
	if err == sql.ErrNoRows {
		qInsert := sq.Insert("pubsub_jids").
			Columns([]string{"jid", "jid_sha1"}...).
			Values([]interface{}{jid, s.Sha1(jid)}...)

		sqlRet, err := qInsert.RunWith(s.db).Exec()
		if err == nil {
			retJidId, _ = sqlRet.LastInsertId()
			return
		}
	}
	retJidId = qJidId
	return
}

func (s *Storage) UpdateNodeConfig(jid jid.JID, nodeId int64, serializedData string, collectionId int64) (affectRows int64) {
	affectRows = 0
	updateRet, err := s.db.Exec(`update pubsub_nodes set configuration = ?, collection_id = ? where node_id = ?`, serializedData, collectionId, nodeId)
	if err != nil {
		affectRows = 0
	}

	affectRows, _ = updateRet.RowsAffected()
	return
}

func (s *Storage) GetNodeMeta(serviceJid jid.JID, nodeName string) (*model.NodeMeta, error) {
	var nodeMetaVar model.NodeMeta
	err := s.db.QueryRow(`
		select n.node_id, n.configuration, cj.jid, n.creation_date
		from pubsub_nodes n
		inner join pubsub_service_jids sj on n.service_id = sj.service_id
		inner join pubsub_jids cj on cj.jid_id = n.creator_id
		where sj.service_jid_sha1 = ? and n.name_sha1 = ?
			and sj.service_jid = ? and n.name = ? `,
		s.Sha1(serviceJid.String()), s.Sha1(nodeName), serviceJid.String(), nodeName).
		Scan(&nodeMetaVar.NodeId, &nodeMetaVar.NodeConfig, &nodeMetaVar.Creator, &nodeMetaVar.CreateDate)
	if err != nil {
		return nil, err
	}

	return &nodeMetaVar, nil
}

func (s *Storage) SetNodeAffiliation(serviceJid jid.JID, nodeId int64, nodeName string, affiliation stateless.UsersAffiliation) (error) {
	var (
		err       error
		vJidId    int64
		vAffExist int64
	)
	jid := affiliation.GetJid().ToBareJID().String()
	err = s.db.QueryRow("select jid_id  from pubsub_jids where jid_sha1 = ? and jid = ?", s.Sha1(jid), jid).Scan(&vJidId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	err = nil

	if vJidId > 0 {
		err = s.db.QueryRow("select 1 from pubsub_affiliations pa where pa.node_id = ? and pa.jid_id = ?", nodeId, vJidId).Scan(vAffExist)
		if err == nil {
			vAffExist = 1
		} else if err == sql.ErrNoRows {
			vAffExist = 0
			err = nil
		} else {
			return err
		}
	}

	if affiliation.GetAffiliation() != enums.AffiliationNone {
		if vJidId <= 0 {
			vJidId = s.PubSubEnsureJid(jid)
		}

		if vAffExist > 0 {
			_, err = s.db.Exec("update pubsub_affiliations set affiliation = ? where node_id = ? and jid_id = ?",
				affiliation.GetAffiliation().String(), nodeId, vJidId)
			if err != nil {
				return err
			}
		} else {
			_, err = s.db.Exec("insert into pubsub_affiliations (node_id, jid_id, affiliation) values (?, ?, ?)",
				nodeId, vJidId, affiliation.GetAffiliation().String())
			if err != nil {
				return err
			}
		}
	} else {
		if vAffExist > 0 {
			_, err = s.db.Exec("delete from pubsub_affiliations where node_id = ? and jid_id = ?", nodeId, vJidId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Storage) SetNodeSubscription(serviceJid jid.JID, nodeId int64, nodeName string, subscription stateless.UsersSubscription) (error) {
	var (
		err       error
		vJidId    int64
		vSubExist int64
	)
	jid := subscription.GetJid().ToBareJID().String()

	vJidId = s.PubSubEnsureJid(jid)
	if vJidId > 0 {
		err = s.db.QueryRow("select 1  from pubsub_subscriptions where node_id = ? and jid_id = ?", nodeId, vJidId).Scan(vSubExist)
		if err == nil {
			vSubExist = 1
		} else if err == sql.ErrNoRows {
			vSubExist = 0
			err = nil
		} else {
			return err
		}
	}

	if vSubExist > 0 {
		_, err = s.db.Exec("update pubsub_subscriptions set subscription = ? where node_id = ? and jid_id = ?",
			subscription.GetSubscription().String(), nodeId, vJidId)
		if err != nil {
			return err
		}
	} else {
		_, err = s.db.Exec("insert into pubsub_subscriptions (node_id,jid_id,subscription,subscription_id) values (?,?,?,?)",
			nodeId, vJidId, subscription.GetSubscription().String(), subscription.GetSubid())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) GetNodeAffiliations(serviceJid jid.JID, nodeId int64) (*cached.NodeAffiliations, error) {
	var err error
	rows, err := s.db.Query(`
		select pj.jid, pa.affiliation from pubsub_affiliations pa
		inner join pubsub_jids pj on pa.jid_id = pj.jid_id
		where pa.node_id = ?`, nodeId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	nodeAffiliations := cached.NewNodeAffiliations()
	for rows.Next() {
		var (
			scanJid string
			scanAff string
		)
		err = rows.Scan(&scanJid, &scanAff)
		if err != nil {
			return nil, err
		}
		saveJid, _ := jid.NewWithString(scanJid, false)
		aff := enums.AffiliationType(scanAff)
		nodeAffiliations.AddAffiliation(*saveJid, aff)
	}
	nodeAffiliations.AffiliationsSaved()
	return nodeAffiliations, nil
}

func (s *Storage) GetNodeSubscriptions(serviceJid jid.JID, nodeId int64) (*cached.NodeSubscriptions, error) {
	var err error
	rows, err := s.db.Query(`
		select pj.jid, ps.subscription, ps.subscription_id
		from pubsub_subscriptions ps
		inner join pubsub_jids pj on ps.jid_id = pj.jid_id
		where ps.node_id = ?;`, nodeId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	nodeSubscriptions := cached.NewNodeSubscriptions()
	for rows.Next() {
		var (
			scanJid   string
			scanSub   string
			scanSubid string
		)
		err = rows.Scan(&scanJid, &scanSub, &scanSubid)
		if err != nil {
			return nil, err
		}
		saveJid, _ := jid.NewWithString(scanJid, false)
		sub := enums.SubscriptionType(scanSub)
		nodeSubscriptions.AddSubscription(*saveJid, sub, scanSubid)
	}
	nodeSubscriptions.SubscriptionsSaved()
	return nodeSubscriptions, nil
}
