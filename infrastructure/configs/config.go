package configs

// Add to existing config.go
func InitializeConfig() {
	LoadEnv()
	InitializeLogger()
}
