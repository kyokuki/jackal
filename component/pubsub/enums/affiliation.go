package enums

type affiliationType string

const (
	AffiliationOwner   = affiliationType("owner")
	AffiliationAdmin   = affiliationType("admin")
	AffiliationMember  = affiliationType("member")
	AffiliationNone    = affiliationType("none")
	AffiliationOutcast = affiliationType("outcast")
)

type affiliationPrivileges struct {
	configureNode bool
	deleteItem    bool
	deleteNode    bool
	publishItem   bool
	purgeNode     bool
	retrieveItem  bool
	subscribe     bool
	weight        int
}

var affiliations struct {
	admin   affiliationPrivileges
	member  affiliationPrivileges
	none    affiliationPrivileges
	outcast affiliationPrivileges
	owner   affiliationPrivileges
}

func init() {
	affiliations.admin = newAffiliationPrivileges(3, true, true, true, true, false, false, false)
	affiliations.member = newAffiliationPrivileges(2, true, true, false, false, false, false, false)
	affiliations.none = newAffiliationPrivileges(1, true, false, false, false, false, false, false)
	affiliations.outcast = newAffiliationPrivileges(0, false, false, false, false, false, false, false)
	affiliations.owner = newAffiliationPrivileges(4, true, true, true, true, true, true, true)
}

func newAffiliationPrivileges(
	weight int,
	subscribe,
	retrieveItem,
	publishItem,
	deleteItem,
	configureNode,
	deleteNode,
	purgeNode bool) affiliationPrivileges {
	x := affiliationPrivileges{
		weight:        weight,
		subscribe:     subscribe,
		retrieveItem:  retrieveItem,
		publishItem:   publishItem,
		deleteItem:    deleteItem,
		configureNode: configureNode,
		deleteNode:    deleteNode,
		purgeNode:     purgeNode,
	}
	return x
}

func (x affiliationType) privileges() affiliationPrivileges {
	switch x {
	case AffiliationAdmin:
		return affiliations.admin
	case AffiliationMember:
		return affiliations.member
	case AffiliationNone:
		return affiliations.none
	case AffiliationOutcast:
		return affiliations.outcast
	case AffiliationOwner:
		return affiliations.owner
	}
	return affiliations.none
}

func (x affiliationType) Weight() int {
	return x.privileges().weight
}

func (x affiliationType) IsSubscribe() bool {
	return x.privileges().subscribe
}

func (x affiliationType) IsRetrieveItem() bool {
	return x.privileges().retrieveItem
}

func (x affiliationType) IsPurgeNode() bool {
	return x.privileges().purgeNode
}

func (x affiliationType) IsPublishItem() bool {
	return x.privileges().publishItem
}

func (x affiliationType) IsDeleteNode() bool {
	return x.privileges().deleteNode
}

func (x affiliationType) IsDeleteItem() bool {
	return x.privileges().deleteItem
}

func (x affiliationType) IsConfigureNode() bool {
	return x.privileges().configureNode
}

func (x affiliationType) String() string {
	return string(x)
}
