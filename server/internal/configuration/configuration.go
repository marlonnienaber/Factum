package configuration

import (
	"os"
	"strings"
)

type Configuration struct {
	Port                 string
	DatabaseUrl          string
	SignUpEmailWhitelist []string
}

func LoadConfiguration() Configuration {
	port := getStringEnvironmentVariable("PORT", ":3001")
	databaseUrl := getStringEnvironmentVariable("DATABASE_URL", "postgres://username:password@localhost:5435/factum_database?sslmode=disable")
	signUpEmailWhitelist := parseSignUpEmailWhitelist(getStringEnvironmentVariable("SIGN_UP_EMAIL_WHITELIST", ""))

	return Configuration{
		Port:                 port,
		DatabaseUrl:          databaseUrl,
		SignUpEmailWhitelist: signUpEmailWhitelist,
	}
}

func getStringEnvironmentVariable(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func parseSignUpEmailWhitelist(whitelistString string) []string {
	if whitelistString == "" {
		return []string{}
	}

	values := strings.Split(whitelistString, ",")
	emailWhitelist := make([]string, 0, len(values))

	for _, value := range values {
		email := strings.TrimSpace(strings.ToLower(value))
		if email != "" {
			emailWhitelist = append(emailWhitelist, email)
		}
	}

	return emailWhitelist
}
