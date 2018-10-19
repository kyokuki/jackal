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