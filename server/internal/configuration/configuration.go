package configuration

import "os"

type Configuration struct {
	Port        string
	DatabaseUrl string
}

func LoadConfiguration() Configuration {
	port := getEnvironmentVariable("PORT", ":3001")
	databaseUrl := getEnvironmentVariable("DATABASE_URL", "postgres://username:password@localhost:5435/factum_database?sslmode=disable")

	return Configuration{
		Port:        port,
		DatabaseUrl: databaseUrl,
	}
}

func getEnvironmentVariable(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
