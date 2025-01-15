package memory

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	WhitelistIDs []int64 `mapstructure:"whitelist_ids"`
	HourlyLimit  int     `mapstructure:"hourly_limit"`
}
