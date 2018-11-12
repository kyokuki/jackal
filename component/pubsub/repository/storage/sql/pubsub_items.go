package sql

import (
	"database/sql"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
)

func (s *Storage) GetItem(serviceJid jid.JID, nodeId int64, itemId string) (model.ItemMeta, error) {
	var err error
	rows, err := s.db.Query(`
		select pn.name, pi.node_id, pi.data, pj.jid, pi.creation_date, pi.update_date
		from pubsub_items pi
		inner join pubsub_jids pj on pj.jid_id = pi.publisher_id
    	inner join pubsub_nodes pn on pn.node_id = pi.node_id
		where pi.node_id = ? and pi.id_sha1 = SHA1(?) and pi.id = ?`,
		nodeId,
		s.Sha1(itemId),
		itemId)
	if err != nil && err != sql.ErrNoRows {
		return model.ItemMeta{}, err
	}
	defer rows.Close()

	var resultItemMeta model.ItemMeta
	for rows.Next() {
		err = rows.Scan(&resultItemMeta.NodeName, &resultItemMeta.NodeId, &resultItemMeta.Data, &resultItemMeta.Jid, &resultItemMeta.CreateDate, &resultItemMeta.UpdateDate)
		if err != nil {
			return model.ItemMeta{}, err
		}
	}
	return resultItemMeta, nil
}
