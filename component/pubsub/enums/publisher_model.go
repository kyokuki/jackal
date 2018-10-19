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