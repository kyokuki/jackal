package enums

type AffiliationType string

const (
	AffiliationOwner   = AffiliationType("owner")
	AffiliationAdmin   = AffiliationType("admin")
	AffiliationMember  = AffiliationType("member")
	AffiliationNone    = AffiliationType("none")
	AffiliationOutcast = AffiliationType("outcast")
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

func (x AffiliationType) privileges() affiliationPrivileges {
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

func (x AffiliationType) Weight() int {
	return x.privileges().weight
}

func (x AffiliationType) IsSubscribe() bool {
	return x.privileges().subscribe
}

func (x AffiliationType) IsRetrieveItem() bool {
	return x.privileges().retrieveItem
}

func (x AffiliationType) IsPurgeNode() bool {
	return x.privileges().purgeNode
}

func (x AffiliationType) IsPublishItem() bool {
	return x.privileges().publishItem
}

func (x AffiliationType) IsDeleteNode() bool {
	return x.privileges().deleteNode
}

func (x AffiliationType) IsDeleteItem() bool {
	return x.privileges().deleteItem
}

func (x AffiliationType) IsConfigureNode() bool {
	return x.privileges().configureNode
}

func (x AffiliationType) String() string {
	return string(x)
}

func NewAffiliationValue(strSubscription string) AffiliationType {
	switch strSubscription {
	case "owner":
		return AffiliationOwner
	case "admin":
		return AffiliationAdmin
	case "member":
		return AffiliationMember
	case "none":
		return AffiliationNone
	case "outcast":
		return AffiliationOutcast
	default:
		return AffiliationNone
	}
}