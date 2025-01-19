package memory

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	WhitelistIDs     []int64 `mapstructure:"whitelist_ids"`
	UserHourlyLimit  int     `mapstructure:"user_hourly_limit"`
	UserDailyLimit   int     `mapstructure:"user_daily_limit"`
	GlobalDailyLimit int     `mapstructure:"global_daily_limit"`
}
