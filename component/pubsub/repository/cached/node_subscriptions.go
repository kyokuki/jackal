package cached

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/repository/stateless"
	"github.com/ortuman/jackal/component/pubsub/enums"
	"sync"
	"github.com/satori/go.uuid"
	"strings"
)

type NodeSubscriptions struct {
	changedSubs map[jid.JID]stateless.UsersSubscription
	subs map[jid.JID]stateless.UsersSubscription
	mutex sync.RWMutex
}

func NewNodeSubscriptions() *NodeSubscriptions {
	a := NodeSubscriptions{}
	a.changedSubs = make(map[jid.JID]stateless.UsersSubscription)
	a.subs = make(map[jid.JID]stateless.UsersSubscription)
	return &a
}

func (ns *NodeSubscriptions) GetChanged() []stateless.UsersSubscription{
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	var userSubs []stateless.UsersSubscription
	for _, changedSub := range ns.changedSubs {
		userSubs = append(userSubs, changedSub)
	}
	return userSubs
}

func (ns *NodeSubscriptions) GetSubscriptions() []stateless.UsersSubscription{
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	var userSubs []stateless.UsersSubscription
	for _, sub := range ns.subs {
		userSubs = append(userSubs, sub)
	}
	for _, changedSub := range ns.changedSubs {
		userSubs = append(userSubs, changedSub)
	}
	return userSubs
}

func (ns *NodeSubscriptions) AddSubscriberJid(bareJid jid.JID, sub enums.SubscriptionType) string {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	u1 := uuid.Must(uuid.NewV1())
	subid := strings.Replace(u1.String(), "-", "", -1)
	userSub := stateless.NewUsersSubscription(bareJid, subid, sub)
	ns.changedSubs[*bareJid.ToBareJID()] = userSub
	delete(ns.subs, *bareJid.ToBareJID())
	return subid
}

func (ns *NodeSubscriptions) AddSubscription(bareJid jid.JID, sub enums.SubscriptionType, subid string) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	userSub := stateless.NewUsersSubscription(bareJid, subid, sub)
	ns.changedSubs[*bareJid.ToBareJID()] = userSub
	delete(ns.subs, *bareJid.ToBareJID())
}

func (ns *NodeSubscriptions) ChangeSubscription(bareJid jid.JID, sub enums.SubscriptionType) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	userSub1, ok1 := ns.changedSubs[*bareJid.ToBareJID()]
	if ok1 {
		userSub1.SetSubscription(sub)
		ns.changedSubs[*bareJid.ToBareJID()] = userSub1
	}

	userSub2, ok2 := ns.subs[*bareJid.ToBareJID()]
	if ok2 {
		userSub2.SetSubscription(sub)
		ns.changedSubs[*bareJid.ToBareJID()] = userSub2
		delete(ns.subs, *bareJid.ToBareJID())
	}
}

func (ns *NodeSubscriptions) GetSubscription(bareJid jid.JID)  enums.SubscriptionType {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	userSub1, ok1 := ns.subs[*bareJid.ToBareJID()]
	if ok1 {
		return userSub1.GetSubscription()
	}
	userSub2, ok2 := ns.changedSubs[*bareJid.ToBareJID()]
	if ok2 {
		return userSub2.GetSubscription()
	}
	return enums.SubscriptionNone
}

func (ns *NodeSubscriptions) GetSubscriptionId(bareJid jid.JID)  string {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	userSub1, ok1 := ns.subs[*bareJid.ToBareJID()]
	if ok1 {
		return userSub1.GetSubid()
	}
	userSub2, ok2 := ns.changedSubs[*bareJid.ToBareJID()]
	if ok2 {
		return userSub2.GetSubid()
	}
	return ""
}

func (ns *NodeSubscriptions) SubscriptionsNeedsWriting() bool {
	return len(ns.changedSubs) > 0
}

func (ns *NodeSubscriptions) SubscriptionsSaved()  {
	ns.mergeSubscriptions()
}

func (ns *NodeSubscriptions) mergeSubscriptions() {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	for key, item := range ns.changedSubs {
		ns.subs[key] = item
	}
	// clear changed subs
	ns.changedSubs = make(map[jid.JID]stateless.UsersSubscription)
}