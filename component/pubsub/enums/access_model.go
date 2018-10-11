package enums

type accessModelType string

const (
	AccessModelAuthorize = accessModelType("authorize")
	AccessModelOpen      = accessModelType("open")
	AccessModelPresence  = accessModelType("presence")
	AccessModelRoster    = accessModelType("roster")
	AccessModelWhitelist = accessModelType("whitelist")
)

func (x accessModelType) String() string {
	return string(x)
}