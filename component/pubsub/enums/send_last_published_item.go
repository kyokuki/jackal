package enums

type sendLastPublishedItemType string

const (
	SendLastPublishedItem_never               = sendLastPublishedItemType("Never")
	SendLastPublishedItem_on_sub              = sendLastPublishedItemType("When a new subscription is processed")
	SendLastPublishedItem_on_sub_and_presence = sendLastPublishedItemType("When a new subscription is processed and whenever a subscriber comes online")
)

func (x sendLastPublishedItemType) String() string {
	return string(x)
}
