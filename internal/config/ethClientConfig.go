package config

type RetryConfig struct {
	Debug            bool
	RetryCount       int
	RetryWaitTime    int
	RetryMaxWaitTime int
}

type EthClientCfg struct {
	RetryConfig *RetryConfig
	// Add any other configuration fields specific to the EthClient here
}
