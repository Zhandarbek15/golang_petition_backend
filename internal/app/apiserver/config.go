package apiserver

type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
}

type AppConfig struct {
	BindAddr string `json:"bind_addr"`
	BindPort string `json:"bind_port"`
	LogLevel string `json:"log_level"`
}

type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"database"`
}

// NewConfig Возвращает конфигураций по умолчанию
func NewConfig() *Config {
	return &Config{
		App: AppConfig{
			BindAddr: "0.0.0.0",
			BindPort: "8080",
			LogLevel: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			Username: "root",
			Password: "root",
		},
	}
}
