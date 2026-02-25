package proxy

// usageInfo represents the token usage information from various AI providers.
// It supports OpenAI, Anthropic, and Gemini style JSON structures through multiple tags.
type usageInfo struct {
	// OpenAI and general standard
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`

	// Some providers use camelCase even for OpenAI fields
	PromptTokensCamel     int `json:"promptTokens"`
	CompletionTokensCamel int `json:"completionTokens"`
	TotalTokensCamel      int `json:"totalTokens"`

	// Anthropic fields
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`

	// Gemini fields (Standard camelCase used by Google)
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`

	// Some Gemini-compatible proxies might use snake_case
	PromptTokenCountSnake     int `json:"prompt_token_count"`
	CandidatesTokenCountSnake int `json:"candidates_token_count"`
	TotalTokenCountSnake      int `json:"total_token_count"`
}

// Normalize ensures that all token fields are correctly populated regardless of source format.
func (u *usageInfo) Normalize() {
	// 1. Merge OpenAI camelCase variations
	if u.PromptTokens == 0 && u.PromptTokensCamel > 0 {
		u.PromptTokens = u.PromptTokensCamel
	}
	if u.CompletionTokens == 0 && u.CompletionTokensCamel > 0 {
		u.CompletionTokens = u.CompletionTokensCamel
	}
	if u.TotalTokens == 0 && u.TotalTokensCamel > 0 {
		u.TotalTokens = u.TotalTokensCamel
	}

	// 2. Anthropic mapping
	if u.PromptTokens == 0 && u.InputTokens > 0 {
		u.PromptTokens = u.InputTokens
	}
	if u.CompletionTokens == 0 && u.OutputTokens > 0 {
		u.CompletionTokens = u.OutputTokens
	}

	// 3. Gemini mapping (Check both camelCase and snake_case)
	if u.PromptTokens == 0 {
		if u.PromptTokenCount > 0 {
			u.PromptTokens = u.PromptTokenCount
		} else if u.PromptTokenCountSnake > 0 {
			u.PromptTokens = u.PromptTokenCountSnake
		}
	}
	if u.CompletionTokens == 0 {
		if u.CandidatesTokenCount > 0 {
			u.CompletionTokens = u.CandidatesTokenCount
		} else if u.CandidatesTokenCountSnake > 0 {
			u.CompletionTokens = u.CandidatesTokenCountSnake
		}
	}
	if u.TotalTokens == 0 {
		if u.TotalTokenCount > 0 {
			u.TotalTokens = u.TotalTokenCount
		} else if u.TotalTokenCountSnake > 0 {
			u.TotalTokens = u.TotalTokenCountSnake
		}
	}

	// Final sum if TotalTokens is still missing
	if u.TotalTokens == 0 {
		u.TotalTokens = u.PromptTokens + u.CompletionTokens
	}
}
