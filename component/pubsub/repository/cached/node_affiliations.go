package cached

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"sync"
)

type NodeAffiliations struct {
	affs map[jid.JID]stateless.UsersAffiliation
	mutex sync.RWMutex
}

func NewNodeAffiliations() *NodeAffiliations {
	a := NodeAffiliations{}
	a.affs = make(map[jid.JID]stateless.UsersAffiliation)
	return &a
}

func (na *NodeAffiliations) GetAffiliations() []stateless.UsersAffiliation{
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	var userAffs []stateless.UsersAffiliation
	for _, aff := range na.affs {
		userAffs = append(userAffs, aff)
	}
	return userAffs
}

func (na *NodeAffiliations) AddAffiliation(bareJid jid.JID, affiliation enums.AffiliationType) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	userAff := stateless.NewUsersAffiliation(bareJid, affiliation)
	na.affs[*bareJid.ToBareJID()] = userAff
}

func (na *NodeAffiliations) ChangeAffiliation(bareJid jid.JID, affiliation enums.AffiliationType) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	userAff, ok := na.affs[*bareJid.ToBareJID()]
	if ok {
		userAff.SetAffiliation(affiliation)
		na.affs[*bareJid.ToBareJID()] = userAff
		return
	}
	userAff = stateless.NewUsersAffiliation(bareJid, enums.AffiliationNone)
	na.affs[*bareJid.ToBareJID()] = userAff
}

func (na *NodeAffiliations) GetSubscriberAffiliation(bareJid jid.JID) stateless.UsersAffiliation{
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	userAff, ok := na.affs[*bareJid.ToBareJID()]
	if ok {
		return userAff
	}
	return stateless.NewUsersAffiliation(bareJid, enums.AffiliationNone)
}

