package peer

type MessageStream struct {
	Data string `json:"data"`
}

type SignerResult struct {
	R string `yaml:"r"`
	S string `yaml:"s"`
}
