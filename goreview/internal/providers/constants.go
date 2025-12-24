package providers

// Common error format strings (SonarQube S1192)
const (
	ErrMarshalRequest = "marshal request: %w"
	ErrCreateRequest  = "create request: %w"
	ErrDecodeResponse = "decode response: %w"
)

// Common HTTP constants
const (
	ContentTypeJSON     = "application/json"
	ChatCompletionsPath = "/chat/completions"
	APIGeneratePath     = "/api/generate"
)
