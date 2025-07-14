package gigago

type GenerativeModel struct {
	c                 *Client
	fullName          string
	SystemInstruction string
	// Nucleus sampling (top-p). Limits token selection to the smallest set whose total probability is ≥ top_p (range: 0.0–1.0). Default: 1
	TopP float64
	// Sampling temperature. Higher values = more randomness, lower = more deterministic output. Default: 0
	Temperature float64
	// Maximum number of tokens allowed in the generated response. Default 999999999.
	MaxTokens int32
	// Penalizes repeated tokens. Values > 1.0 discourage repetition (1.0 = no penalty). Default 1
	RepetitionPenalty float64
}

// GenerativeModel returns a new GenerativeModel instance for the specified model name (e.g., "GigaChat").
// The returned model can be configured by setting its fields (e.g., Temperature, TopP)
// before being used to generate content.
func (c *Client) GenerativeModel(name string) *GenerativeModel {
	return &GenerativeModel{
		c:                 c,
		fullName:          name,
		SystemInstruction: "",
		Temperature:       0,
		TopP:              1,
		MaxTokens:         999999999,
		RepetitionPenalty: 1,
	}
}
