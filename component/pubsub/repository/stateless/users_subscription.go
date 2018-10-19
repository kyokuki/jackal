package stateless

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type UsersSubscription struct {
	jid jid.JID
	subid string
	subscription enums.SubscriptionType
}

func NewUsersSubscription(jid jid.JID, subid string, subscription enums.SubscriptionType) UsersSubscription {
	s := UsersSubscription{
		jid : *jid.ToBareJID(),
		subid: subid,
		subscription: subscription,
	}
	return s
}

func (us *UsersSubscription) GetJid() jid.JID {
	return us.jid
}

func (us *UsersSubscription) GetSubid() string {
	return us.subid
}

func (us *UsersSubscription) GetSubscription() enums.SubscriptionType {
	return us.subscription
}

func (us *UsersSubscription) SetSubscription(subscription enums.SubscriptionType) {
	us.subscription = subscription
}


