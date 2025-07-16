package gigago

import "fmt"

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
	if name == "" {
		name = "GigaChat" // Default model name
	}

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

// Validate checks if the model parameters are within acceptable ranges
func (g *GenerativeModel) Validate() error {
	if g.Temperature < 0 || g.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got %f", g.Temperature)
	}
	if g.TopP < 0 || g.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1, got %f", g.TopP)
	}
	if g.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive, got %d", g.MaxTokens)
	}
	if g.RepetitionPenalty < 0.1 || g.RepetitionPenalty > 2.0 {
		return fmt.Errorf("repetition_penalty must be between 0.1 and 2.0, got %f", g.RepetitionPenalty)
	}
	return nil
}
