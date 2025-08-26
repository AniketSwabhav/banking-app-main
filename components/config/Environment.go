package config

type Environment string

type EnvKey string

// GetStringValue will fetch the env's value from file and return it as string.
func (e EnvKey) GetStringValue() string {
	return GlobalConfig.GetString(e)
}

func (e EnvKey) GetInt64Value() int64 {
	return GlobalConfig.GetInt64(e)
}
