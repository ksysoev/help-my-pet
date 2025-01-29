package i18n

// Message represents a message type
type Message string

const (
	ErrorMessage       Message = "error"
	StartMessage       Message = "start"
	RateLimitMessage   Message = "rate_limit"
	GlobalLimitMessage Message = "global_limit"
	MessageTooLong     Message = "message_too_long"
)

var defaultMessages = Messages{
	Error:          "Sorry, I encountered an error while processing your request. Please try again later.",
	Start:          "Welcome to Help My Pet Bot! How can I help you today?",
	RateLimit:      "You have reached the maximum number of requests per hour. Please try again later.",
	GlobalLimit:    "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
	MessageTooLong: "I apologize, but your message is too long for me to process. Please try to make it shorter and more concise.",
}

// Config holds translations for all supported languages
type Config struct {
	Languages map[string]Messages `mapstructure:"languages"`
}

// Messages holds translations for a specific language
type Messages struct {
	Error          string `mapstructure:"error"`
	Start          string `mapstructure:"start"`
	RateLimit      string `mapstructure:"rate_limit"`
	GlobalLimit    string `mapstructure:"global_limit"`
	MessageTooLong string `mapstructure:"message_too_long"`
}

// GetMessage returns a translated message for the given language and message type
func (c *Config) GetMessage(lang string, msgType Message) string {
	if lang == "" {
		lang = "en"
	}

	var msg Messages

	if m, ok := c.Languages[lang]; ok {
		msg = m
	} else if m, ok := c.Languages["en"]; ok {
		msg = m
	} else {
		msg = defaultMessages
	}

	switch msgType {
	case ErrorMessage:
		return msg.Error
	case StartMessage:
		return msg.Start
	case RateLimitMessage:
		return msg.RateLimit
	case GlobalLimitMessage:
		return msg.GlobalLimit
	case MessageTooLong:
		return msg.MessageTooLong
	default:
		return "An error occurred."
	}
}
