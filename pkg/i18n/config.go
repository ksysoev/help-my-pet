package i18n

// Message represents a message type
type Message string

const (
	ErrorMessage     Message = "error"
	StartMessage     Message = "start"
	RateLimitMessage Message = "rate_limit"
)

// Config holds translations for all supported languages
type Config struct {
	Languages map[string]Messages `mapstructure:"languages"`
}

// Messages holds translations for a specific language
type Messages struct {
	Error     string `mapstructure:"error"`
	Start     string `mapstructure:"start"`
	RateLimit string `mapstructure:"rate_limit"`
}

// GetMessage returns a translated message for the given language and message type
func (c *Config) GetMessage(lang string, msgType Message) string {
	if lang == "" {
		lang = "en"
	}

	if messages, ok := c.Languages[lang]; ok {
		switch msgType {
		case ErrorMessage:
			return messages.Error
		case StartMessage:
			return messages.Start
		case RateLimitMessage:
			return messages.RateLimit
		}
	}

	// Fallback to English if translation not found
	if messages, ok := c.Languages["en"]; ok {
		switch msgType {
		case ErrorMessage:
			return messages.Error
		case StartMessage:
			return messages.Start
		case RateLimitMessage:
			return messages.RateLimit
		}
	}

	// Default fallback messages in case config is not properly loaded
	switch msgType {
	case ErrorMessage:
		return "Sorry, I encountered an error while processing your request. Please try again later."
	case StartMessage:
		return "Welcome to Help My Pet Bot! How can I help you today?"
	case RateLimitMessage:
		return "You have reached the maximum number of requests per hour. Please try again later."
	default:
		return "An error occurred."
	}
}
