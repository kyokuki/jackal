package sql

import (
	sq "github.com/Masterminds/squirrel"
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
	"time"
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
		values := []interface{}{serviceId, nodeName, s.Sha1(nodeName), nodeType, jidId, time.Now(),serializedNodeConfig, collection}

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
			and sj.service_jid = ? and n.name = ?;
`, s.Sha1(serviceJid.String()), s.Sha1(nodeName), serviceJid.String(), nodeName).Scan(&nodeMetaVar.NodeId, &nodeMetaVar.NodeConfig, &nodeMetaVar.Creator, &nodeMetaVar.CreateDate)
	if err != nil {
		return nil, err
	}

	return &nodeMetaVar, nil
}
