package enums

type SendLastPublishedItemType string

const (
	SendLastPublishedItem_never               = SendLastPublishedItemType("Never")
	SendLastPublishedItem_on_sub              = SendLastPublishedItemType("When a new subscription is processed")
	SendLastPublishedItem_on_sub_and_presence = SendLastPublishedItemType("When a new subscription is processed and whenever a subscriber comes online")
)

func (x SendLastPublishedItemType) String() string {
	return string(x)
}
