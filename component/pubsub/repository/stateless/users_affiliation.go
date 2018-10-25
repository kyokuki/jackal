package stateless

import (
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/component/pubsub/enums"
)

type UsersAffiliation struct {
	jid jid.JID
	affiliation enums.AffiliationType
}

func NewUsersAffiliation(jid jid.JID, affiliation enums.AffiliationType) UsersAffiliation {
	u := UsersAffiliation{
		jid : *jid.ToBareJID(),
		affiliation: affiliation,
	}
	return u
}

func (ua *UsersAffiliation) GetJid() *jid.JID {
	return &ua.jid
}

func (ua *UsersAffiliation) GetAffiliation() enums.AffiliationType {
	return ua.affiliation
}

func (ua *UsersAffiliation) SetAffiliation(affiliation enums.AffiliationType) {
	ua.affiliation = affiliation
}


