package sql

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"database/sql"
	"github.com/ortuman/jackal/component/pubsub/repository/storage/model"
)

func (s *Storage) GetUserRoster(owner jid.JID) ([]model.UserRosterItem, error) {
	var err error

	rows, err := s.db.Query(`
		select username, jid, name, subscription
		from roster_items
		where username = ? `, owner.Node())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var userRosterArr []model.UserRosterItem
	for rows.Next() {
		var userRoster model.UserRosterItem
		err = rows.Scan(&userRoster.Username, &userRoster.JID, &userRoster.Subscription)
		if err != nil {
			return nil, err
		}
		userRosterArr = append(userRosterArr, userRoster)
	}
	return userRosterArr, nil
}