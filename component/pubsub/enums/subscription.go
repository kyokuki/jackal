package enums

type SubscriptionType string

const (
	SubscriptionNone = SubscriptionType("none")
	SubscriptionPending = SubscriptionType("pending")
	SubscriptionSubscribed = SubscriptionType("subscribed")
	SubscriptionUnconfigured = SubscriptionType("unconfigured")
)

func (x SubscriptionType) String() string {
	return string(x)
}

func NewSubscriptionValue(strSubscription string) SubscriptionType {
	switch strSubscription {
	case "none":
		return SubscriptionNone
	case "pending":
		return SubscriptionPending
	case "subscribed":
		return SubscriptionSubscribed
	case "unconfigured":
		return SubscriptionUnconfigured
	default:
		return SubscriptionNone
	}
}