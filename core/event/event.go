package event

type ProviderWebhookEvent struct {
	Name    string
	Headers map[string]string
	Body    []byte
}
