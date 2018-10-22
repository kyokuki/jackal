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