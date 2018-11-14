package sql

import (
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
	"fmt"
)

func (s *Storage) GetItem(serviceJid jid.JID, nodeId int64, itemId string) (model.ItemMeta, error) {
	var err error
	rows, err := s.db.Query(`
		select pn.name, pi.node_id, pi.id, pi.data, pj.jid, pi.creation_date, pi.update_date
		from pubsub_items pi
		inner join pubsub_jids pj on pj.jid_id = pi.publisher_id
    	inner join pubsub_nodes pn on pn.node_id = pi.node_id
		where pi.node_id = ? and pi.id_sha1 = ? and pi.id = ?`,
		nodeId,
		s.Sha1(itemId),
		itemId)
	if err != nil && err != sql.ErrNoRows {
		return model.ItemMeta{}, err
	}
	defer rows.Close()

	var resultItemMeta model.ItemMeta
	for rows.Next() {
		err = rows.Scan(&resultItemMeta.NodeName, &resultItemMeta.NodeId, &resultItemMeta.Id, &resultItemMeta.Data,
			&resultItemMeta.Jid, &resultItemMeta.CreateDate, &resultItemMeta.UpdateDate)
		if err != nil {
			return model.ItemMeta{}, err
		}
	}
	return resultItemMeta, nil
}

func (s *Storage) QueryItems(nodeId int64, orderDate bool, orderAsc bool, limit int64) ([]model.ItemMeta, error) {
	var err error
	querySql := `
		select pn.name, pi.node_id, pi.id, pi.data, pj.jid, pi.creation_date, pi.update_date
		from pubsub_items pi
		inner join pubsub_jids pj on pj.jid_id = pi.publisher_id
    	inner join pubsub_nodes pn on pn.node_id = pi.node_id
		where pi.node_id = ? `

	var rows *sql.Rows

	if orderDate {
		querySql = querySql + "order by update_date"
	} else {
		querySql = querySql + "order by creation_date"
	}

	if orderAsc {
		querySql = querySql + " asc "
	} else {
		querySql = querySql + " desc "
	}

	if limit > 0 {
		querySql = querySql + fmt.Sprintf(" limit %d ", limit)
	}

	rows, err = s.db.Query(querySql, nodeId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var resultItemMetaArr []model.ItemMeta
	for rows.Next() {
		var resultItemMeta model.ItemMeta
		err = rows.Scan(&resultItemMeta.NodeName, &resultItemMeta.NodeId, &resultItemMeta.Id, &resultItemMeta.Data,
			&resultItemMeta.Jid, &resultItemMeta.CreateDate, &resultItemMeta.UpdateDate)
		if err != nil {
			return nil, err
		}
		resultItemMetaArr = append(resultItemMetaArr, resultItemMeta)
	}
	return resultItemMetaArr, nil
}

func (s *Storage) WriteItem(serviceJid jid.JID, nodeId int64, nodeName string, itemId string, publisherJid jid.JID, itemData string) (error) {
	var (
		err    error
		vJidId int64
	)

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

	vJidId, err = s.privatePubSubEnsureJid(tx, publisherJid.ToBareJID().String())
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		insert into pubsub_items (node_id, id_sha1, id, creation_date, update_date, publisher_id, data)
		values (?, ?, ?, UTC_TIMESTAMP(), UTC_TIMESTAMP(), ?, ?)`,
		nodeId, s.Sha1(itemId), itemId, vJidId, itemData)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (s *Storage) DeleteItem(serviceJid jid.JID, nodeId int64, itemId string) (error) {
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

	_, err = tx.Exec("delete from pubsub_items where node_id = ? and id_sha1 = ? and id = ?", nodeId, s.Sha1(itemId), itemId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (s *Storage) GetItemIds(nodeId int64) ([]string, error) {
	var err error
	querySql := `
		select pn.name, pi.node_id, pi.id, pi.data, pj.jid, pi.creation_date, pi.update_date
		from pubsub_items pi
		inner join pubsub_jids pj on pj.jid_id = pi.publisher_id
    	inner join pubsub_nodes pn on pn.node_id = pi.node_id
		where pi.node_id = ? `

	var rows *sql.Rows
	rows, err = s.db.Query(querySql, nodeId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var idsArr []string
	for rows.Next() {
		var resultItemMeta model.ItemMeta
		err = rows.Scan(&resultItemMeta.NodeName, &resultItemMeta.NodeId, &resultItemMeta.Id, &resultItemMeta.Data,
			&resultItemMeta.Jid, &resultItemMeta.CreateDate, &resultItemMeta.UpdateDate)
		if err != nil {
			return nil, err
		}
		idsArr = append(idsArr, resultItemMeta.Id)
	}
	return idsArr, nil
}