package enums

type AccessModelType string

const (
	AccessModelAuthorize = AccessModelType("authorize")
	AccessModelOpen      = AccessModelType("open")
	AccessModelPresence  = AccessModelType("presence")
	AccessModelRoster    = AccessModelType("roster")
	AccessModelWhitelist = AccessModelType("whitelist")
)

func (x AccessModelType) String() string {
	return string(x)
}

func NewAccessModelType(strAccess string) AccessModelType {
	switch strAccess {
	case "authorize":
		return AccessModelAuthorize
	case "open":
		return AccessModelOpen
	case "presence":
		return AccessModelPresence
	case "roster":
		return AccessModelRoster
	case "whitelist":
		return AccessModelWhitelist
	default:
		return AccessModelType("")
	}
}