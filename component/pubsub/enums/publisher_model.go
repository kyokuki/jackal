package enums

type publisherModelType string

const (
	PublisherModelOpen        = publisherModelType("open")
	PublisherModelPublishers  = publisherModelType("publishers")
	PublisherModelSubscribers = publisherModelType("subscribers")
)

func (x publisherModelType) String() string {
	return string(x)
}