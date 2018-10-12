package enums

type subscriptionType string

const (
	SubscriptionNone = subscriptionType("none")
	SubscriptionPending = subscriptionType("pending")
	SubscriptionSubscribed = subscriptionType("subscribed")
	SubscriptionUnconfigured = subscriptionType("unconfigured")
)

func (x subscriptionType) String() string {
	return string(x)
}