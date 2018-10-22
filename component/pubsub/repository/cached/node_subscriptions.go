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
	subs map[jid.JID]stateless.UsersSubscription
	mutex sync.RWMutex
}

func NewNodeSubscriptions() *NodeSubscriptions {
	a := NodeSubscriptions{}
	a.subs = make(map[jid.JID]stateless.UsersSubscription)
	return &a
}

func (ns *NodeSubscriptions) GetSubscriptions() []stateless.UsersSubscription{
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	var userSubs []stateless.UsersSubscription
	for _, sub := range ns.subs {
		userSubs = append(userSubs, sub)
	}
	return userSubs
}

func (ns *NodeSubscriptions) AddSubscriberJid(bareJid jid.JID, sub enums.SubscriptionType) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	u1 := uuid.Must(uuid.NewV1())
	subid := strings.Replace(u1.String(), "-", "", -1)
	userSub := stateless.NewUsersSubscription(bareJid, subid, sub)
	ns.subs[*bareJid.ToBareJID()] = userSub
}

func (ns *NodeSubscriptions) ChangeSubscription(bareJid jid.JID, sub enums.SubscriptionType) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	userSub, ok := ns.subs[*bareJid.ToBareJID()]
	if ok {
		userSub.SetSubscription(sub)
		ns.subs[*bareJid.ToBareJID()] = userSub
	}
}

func (ns *NodeSubscriptions) GetSubscription(bareJid jid.JID)  enums.SubscriptionType {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	userSub, ok := ns.subs[*bareJid.ToBareJID()]
	if ok {
		return userSub.GetSubscription()
	}
	return enums.SubscriptionNone
}

func (ns *NodeSubscriptions) GetSubscriptionId(bareJid jid.JID)  string {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()
	userSub, ok := ns.subs[*bareJid.ToBareJID()]
	if ok {
		return userSub.GetSubid()
	}
	return ""
}

