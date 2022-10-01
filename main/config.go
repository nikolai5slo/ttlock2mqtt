package main

type Config struct {
	Server struct {
		Address string `env:"SERVER_ADDRESS" env-default:"0.0.0.0:8080"`
	}
	TTLock struct {
		Server       string `env:"TTLOCK_SERVER" env-default:"https://euapi.ttlock.com/"`
		ClientID     string `env:"TTLOCK_CLIENT_ID"`
		ClientSecret string `env:"TTLOCK_CLIENT_SECRET"`
	}
	Storage struct {
		FilePath string `env:"STORAGE_FILE" env-default:"./storage.json"`
	}
	Mqtt struct {
		//""
		Broker   string `env:"MQTT_BROKER"`
		ClientID string `env:"MQTT_CLIENT_ID" env-default:"ttlock2mqtt"`
		Username string `env:"MQTT_USERNAME"`
		Password string `env:"MQTT_PASSWORD"`
	}
}
