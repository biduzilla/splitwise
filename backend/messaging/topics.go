package messaging

type Topic string

const (
	RateioEvents Topic = "rateio.events"
)

func (t Topic) String() string { return string(t) }
