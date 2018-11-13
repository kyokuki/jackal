package enums

type PublisherModelType string

const (
	PublisherModelOpen        = PublisherModelType("open")
	PublisherModelPublishers  = PublisherModelType("publishers")
	PublisherModelSubscribers = PublisherModelType("subscribers")
)

func (x PublisherModelType) String() string {
	return string(x)
}

func NewPublisherModelType(strPublisherModel string) PublisherModelType {
	switch strPublisherModel {
	case "open":
		return PublisherModelOpen
	case "publishers":
		return PublisherModelPublishers
	case "subscribers":
		return PublisherModelSubscribers
	default:
		return PublisherModelType("")
	}
}