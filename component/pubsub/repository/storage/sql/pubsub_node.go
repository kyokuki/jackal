package sql

import (
	sq "github.com/Masterminds/squirrel"
	"database/sql"
	"github.com/ortuman/jackal/component/pubsub/model"
)

func (s *Storage) GetNodeConfig(serviceJid string, nodeName string) (string, error) {

	pubsubNode, err := s.GetNode(serviceJid, nodeName)
	if pubsubNode != nil {
		return pubsubNode.Configuration, nil
	}
	return "",err
}

func (s *Storage) GetNode(serviceJid string, nodeName string) (*model.PubSubNode, error) {

	q := sq.Select("service_jid", "name", "node_type", "title", "description", "creator_jid", "created_at", "configuration", "collection_jid").
		From("pubsub_nodes").
		Where(sq.Eq{
		"service_jid": serviceJid,
		"name" : nodeName,
	}).OrderBy("node_id")

	pubsubNode := model.PubSubNode{}
	err := q.RunWith(s.db).QueryRow().Scan(&pubsubNode.ServiceJid,
		&pubsubNode.Name,
		&pubsubNode.NodeType,
		&pubsubNode.Title,
		&pubsubNode.Description,
		&pubsubNode.CreatorJid,
		&pubsubNode.CreatedAt,
		&pubsubNode.Configuration,
		&pubsubNode.CollectionJId)
	switch err {
	case nil:
		return &pubsubNode, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}

	return nil, nil
}