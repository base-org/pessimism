package client

type PagerdutyConfig struct {
	IntegrationKey  string
	ChangeEventsURL string
	AlertEventsURL  string
}

type PagerdutyClient interface {
}

type pagerdutyClient struct {
	integrationKey  string
	changeEventsURL string
	alertEventsURL  string
}

func NewPagerdutyClient(cfg *PagerdutyConfig) PagerdutyClient {
	return &pagerdutyClient{
		integrationKey:  cfg.IntegrationKey,
		changeEventsURL: cfg.ChangeEventsURL,
		alertEventsURL:  cfg.AlertEventsURL,
	}
}
