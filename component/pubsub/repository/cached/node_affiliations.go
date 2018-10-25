package cached

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"sync"
)

type NodeAffiliations struct {
	changedAffs map[jid.JID]stateless.UsersAffiliation
	affs map[jid.JID]stateless.UsersAffiliation
	mutex sync.RWMutex
}

func NewNodeAffiliations() *NodeAffiliations {
	a := NodeAffiliations{}
	a.changedAffs = make(map[jid.JID]stateless.UsersAffiliation)
	a.affs = make(map[jid.JID]stateless.UsersAffiliation)
	return &a
}

func (na *NodeAffiliations) GetChanged() []stateless.UsersAffiliation{
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	var userAffs []stateless.UsersAffiliation
	for _, changedAffs := range na.changedAffs {
		userAffs = append(userAffs, changedAffs)
	}
	return userAffs
}

func (na *NodeAffiliations) GetAffiliations() []stateless.UsersAffiliation{
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	var userAffs []stateless.UsersAffiliation
	for _, aff := range na.affs {
		userAffs = append(userAffs, aff)
	}
	for _, changedAffs := range na.changedAffs {
		userAffs = append(userAffs, changedAffs)
	}
	return userAffs
}

func (na *NodeAffiliations) AddAffiliation(bareJid jid.JID, affiliation enums.AffiliationType) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	userAff := stateless.NewUsersAffiliation(bareJid, affiliation)
	na.changedAffs[*bareJid.ToBareJID()] = userAff
	delete(na.affs, *bareJid.ToBareJID())
}

func (na *NodeAffiliations) ChangeAffiliation(bareJid jid.JID, affiliation enums.AffiliationType) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	userAff1, ok1 := na.changedAffs[*bareJid.ToBareJID()]
	if ok1 {
		userAff1.SetAffiliation(affiliation)
		na.changedAffs[*bareJid.ToBareJID()] = userAff1
		return
	}

	userAff2, ok2 := na.affs[*bareJid.ToBareJID()]
	if ok2 {
		userAff2.SetAffiliation(affiliation)
		na.changedAffs[*bareJid.ToBareJID()] = userAff2
		delete(na.affs, *bareJid.ToBareJID())
		return
	}

	userAff3 := stateless.NewUsersAffiliation(bareJid, enums.AffiliationNone)
	na.changedAffs[*bareJid.ToBareJID()] = userAff3
}

func (na *NodeAffiliations) GetSubscriberAffiliation(bareJid jid.JID) stateless.UsersAffiliation{
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	userAff1, ok1 := na.affs[*bareJid.ToBareJID()]
	if ok1 {
		return userAff1
	}

	userAff2, ok2 := na.changedAffs[*bareJid.ToBareJID()]
	if ok2 {
		return userAff2
	}
	return stateless.NewUsersAffiliation(bareJid, enums.AffiliationNone)
}

func (na *NodeAffiliations) AffiliationsNeedsWriting() bool {
	return len(na.changedAffs) > 0
}

func (na *NodeAffiliations) AffiliationsSaved()  {
	na.mergeAffiliations()
}

func (na *NodeAffiliations) mergeAffiliations() {
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	for key, item := range na.changedAffs {
		na.affs[key] = item
	}
	// clear changed affs
	na.changedAffs = make(map[jid.JID]stateless.UsersAffiliation)
}

