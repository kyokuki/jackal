package sql

import (
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"github.com/ortuman/jackal/component/pubsub/repository/cached"
	"time"
)

func (s *Storage) CreateNode(serviceJid jid.JID, nodeName string, ownerJid jid.JID, nodeConfig base.AbstractNodeConfig, nodeType int, collection string) (int64, error) {
	var err error
	var retNodeId int64 = -1

	serializedNodeConfig := ""
	if nodeConfig != nil {
		serializedNodeConfig = nodeConfig.Form().Element().String()
	}

	tx, err := s.db.Begin()
	if err != nil {
		return retNodeId, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	serviceId, err := s.privatePubSubEnsureServiceJid(tx, serviceJid.String())
	if err != nil {
		return retNodeId, err
	}
	jidId, err := s.privatePubSubEnsureJid(tx, ownerJid.String())
	if err != nil {
		return retNodeId, err
	}

	err = tx.QueryRow("select node_id  from pubsub_nodes where name = ? and service_id = ?", nodeName, serviceId).Scan(&retNodeId)
	if err == sql.ErrNoRows {
		err = nil
		sqlRet, err := tx.Exec("insert into pubsub_nodes (service_id,name,name_sha1,`type`,creator_id,configuration,collection_id,creation_date) values (?, ?, ?, ?, ?, ?, ?, ?)",
			serviceId, nodeName, s.Sha1(nodeName), nodeType, jidId, serializedNodeConfig, collection, time.Now().UTC())
		if err != nil {
			return retNodeId, err
		}
		retNodeId, err = sqlRet.LastInsertId()
	}

	tx.Commit()
	return retNodeId, err
}

func (s *Storage) GetNodeId(serviceJid jid.JID, nodeName string) (int64) {
	var retNodeId int64 = -1
	err := s.db.QueryRow(`
				select n.node_id from pubsub_nodes n 
				inner join pubsub_service_jids sj on n.service_id = sj.service_id
				where sj.service_jid = ? and n.name = ?`, serviceJid.String(), nodeName).Scan(&retNodeId)
	if err != nil {
		retNodeId = -1
	}
	return retNodeId
}

func (s *Storage) privatePubSubEnsureServiceJid(tx *sql.Tx, serviceJid string) (int64, error) {
	var (
		err          error
		retServiceId int64 = -1
		qServiceId   int64
	)

	err = tx.QueryRow("select service_id from pubsub_service_jids where service_jid_sha1 = ? and service_jid = ?",
		s.Sha1(serviceJid), serviceJid).Scan(&qServiceId)
	if err == nil {
		retServiceId = qServiceId
	} else if err == sql.ErrNoRows {
		err = nil
		sqlRet, err := tx.Exec(`insert into pubsub_service_jids (service_jid, service_jid_sha1) values (?, ?)`, serviceJid, s.Sha1(serviceJid))
		if err == nil {
			retServiceId, _ = sqlRet.LastInsertId()
			return retServiceId, err
		}
	}
	return retServiceId, err
}

func (s *Storage) privatePubSubEnsureJid(tx *sql.Tx, jid string) (int64, error) {
	var (
		err      error
		retJidId int64 = -1
		qJidId   int64
	)

	err = tx.QueryRow("select jid_id from pubsub_jids where jid = ? and jid_sha1 = ?", jid, s.Sha1(jid)).Scan(&qJidId)
	if err == nil {
		retJidId = qJidId
	} else if err == sql.ErrNoRows {
		err = nil
		sqlRet, err := tx.Exec("insert into pubsub_jids (jid, jid_sha1) values (?, ?)", jid, s.Sha1(jid))
		if err == nil {
			retJidId, _ = sqlRet.LastInsertId()
			return retJidId, err
		}
	}
	return retJidId, err
}

func (s *Storage) UpdateNodeConfig(jid jid.JID, nodeId int64, serializedData string, collectionId int64) (int64, error) {
	var (
		affectRows int64 = 0
		err        error
	)

	updateRet, err := s.db.Exec(`update pubsub_nodes set configuration = ?, collection_id = ? where node_id = ?`, serializedData, collectionId, nodeId)
	if err != nil {
		affectRows = 0
	}

	affectRows, err = updateRet.RowsAffected()
	return affectRows, err
}

func (s *Storage) GetNodeMeta(serviceJid jid.JID, nodeName string) (*model.NodeMeta, error) {
	var nodeMetaVar model.NodeMeta
	err := s.db.QueryRow(`
		select n.node_id, n.name, n.configuration, cj.jid, n.creation_date
		from pubsub_nodes n
		inner join pubsub_service_jids sj on n.service_id = sj.service_id
		inner join pubsub_jids cj on cj.jid_id = n.creator_id
		where sj.service_jid_sha1 = ? and n.name_sha1 = ?
			and sj.service_jid = ? and n.name = ? `,
		s.Sha1(serviceJid.String()), s.Sha1(nodeName), serviceJid.String(), nodeName).
		Scan(&nodeMetaVar.NodeId, &nodeMetaVar.Name, &nodeMetaVar.NodeConfig, &nodeMetaVar.Creator, &nodeMetaVar.CreateDate)
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	err = tx.QueryRow("select jid_id  from pubsub_jids where jid_sha1 = ? and jid = ?", s.Sha1(jid), jid).Scan(&vJidId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	err = nil

	if vJidId > 0 {
		err = tx.QueryRow("select 1 from pubsub_affiliations pa where pa.node_id = ? and pa.jid_id = ?", nodeId, vJidId).Scan(&vAffExist)
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
			vJidId, err = s.privatePubSubEnsureJid(tx, jid)
			if err != nil {
				return err
			}
		}

		if vAffExist > 0 {
			_, err = tx.Exec("update pubsub_affiliations set affiliation = ? where node_id = ? and jid_id = ?",
				affiliation.GetAffiliation().String(), nodeId, vJidId)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec("insert into pubsub_affiliations (node_id, jid_id, affiliation) values (?, ?, ?)",
				nodeId, vJidId, affiliation.GetAffiliation().String())
			if err != nil {
				return err
			}
		}
	} else {
		if vAffExist > 0 {
			_, err = tx.Exec("delete from pubsub_affiliations where node_id = ? and jid_id = ?", nodeId, vJidId)
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()
	return nil
}

func (s *Storage) SetNodeSubscription(serviceJid jid.JID, nodeId int64, nodeName string, subscription stateless.UsersSubscription) (error) {
	var (
		err       error
		vJidId    int64
		vSubExist int64
	)
	jid := subscription.GetJid().ToBareJID().String()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	vJidId, err = s.privatePubSubEnsureJid(tx, jid)
	if err != nil {
		return err
	}
	if vJidId > 0 {
		err = tx.QueryRow("select 1  from pubsub_subscriptions where node_id = ? and jid_id = ?", nodeId, vJidId).Scan(&vSubExist)
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
		_, err = tx.Exec("update pubsub_subscriptions set subscription = ? where node_id = ? and jid_id = ?",
			subscription.GetSubscription().String(), nodeId, vJidId)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec("insert into pubsub_subscriptions (node_id,jid_id,subscription,subscription_id) values (?,?,?,?)",
			nodeId, vJidId, subscription.GetSubscription().String(), subscription.GetSubid())
		if err != nil {
			return err
		}
	}

	tx.Commit()
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
		aff := enums.NewAffiliationValue(scanAff)
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
		sub := enums.NewSubscriptionValue(scanSub)
		nodeSubscriptions.AddSubscription(*saveJid, sub, scanSubid)
	}
	nodeSubscriptions.SubscriptionsSaved()
	return nodeSubscriptions, nil
}

func (s *Storage) DeleteNode(serviceJid jid.JID, nodeId int64) (error) {
	var err error
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	_, err = tx.Exec("delete from pubsub_items where node_id = ?", nodeId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from pubsub_subscriptions where node_id = ?", nodeId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from pubsub_affiliations where node_id = ?", nodeId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from pubsub_nodes where node_id = ?", nodeId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (s *Storage) GetChildNodes(serviceJid jid.JID, nodeName string) ([]string, error) {
	if nodeName == "" {
		return s.privateGetRootNodes(serviceJid)
	}
	return s.privateGetChildNodes(serviceJid, nodeName)
}

func (s *Storage) privateGetChildNodes(serviceJid jid.JID, nodeName string) ([]string, error) {
	var err error

	rows, err := s.db.Query(`
		select n.node_id, n.name, n.configuration, cj.jid, n.creation_date
		from pubsub_nodes n
		inner join pubsub_service_jids sj on n.service_id = sj.service_id
		inner join pubsub_nodes p on p.node_id = n.collection_id and p.service_id = sj.service_id
		inner join pubsub_jids cj on cj.jid_id = n.creator_id
		where sj.service_jid = ? and p.name = ?`, serviceJid.ToBareJID().String(), nodeName)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var nodeArr []string
	for rows.Next() {
		var nodeMeta model.NodeMeta
		err = rows.Scan(&nodeMeta.NodeId, &nodeMeta.Name, &nodeMeta.NodeConfig, &nodeMeta.Creator, &nodeMeta.CreateDate)
		if err != nil {
			return nil, err
		}
		nodeArr = append(nodeArr, nodeMeta.Name)
	}
	return nodeArr, nil
}

func (s *Storage) privateGetRootNodes(serviceJid jid.JID) ([]string, error) {
	var err error

	rows, err := s.db.Query(`
		select n.node_id, n.name, n.configuration, cj.jid, n.creation_date
		from pubsub_nodes n
		inner join pubsub_service_jids sj on n.service_id = sj.service_id
		inner join pubsub_jids cj on cj.jid_id = n.creator_id
		where sj.service_jid = ? and (n.collection_id is null or n.collection_id = 0)`, serviceJid.ToBareJID().String())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var nodeArr []string
	for rows.Next() {
		var nodeMeta model.NodeMeta
		err = rows.Scan(&nodeMeta.NodeId, &nodeMeta.Name, &nodeMeta.NodeConfig, &nodeMeta.Creator, &nodeMeta.CreateDate)
		if err != nil {
			return nil, err
		}
		nodeArr = append(nodeArr, nodeMeta.Name)
	}
	return nodeArr, nil
}