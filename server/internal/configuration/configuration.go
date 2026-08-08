package configuration

import (
	"os"
	"strings"
)

type Configuration struct {
	Port                 string
	DatabaseUrl          string
	SignUpEmailWhitelist []string
	SessionCookieSecure  bool
	SessionSecretKey     string
}

func LoadConfiguration() Configuration {
	port := getStringEnvironmentVariable("SERVER_PORT", "3001")
	databaseUrl := getStringEnvironmentVariable("DATABASE_URL", "postgres://username:password@localhost:5435/factum_database?sslmode=disable")
	signUpEmailWhitelist := parseSignUpEmailWhitelist(getStringEnvironmentVariable("SIGN_UP_EMAIL_WHITELIST", ""))
	sessionCookieSecure := getBooleanEnvironmentVariable("SESSION_COOKIE_SECURE", false)
	sessionSecretKey := getStringEnvironmentVariable("SESSION_SECRET_KEY", "dev-session-secret-key")

	return Configuration{
		Port:                 port,
		DatabaseUrl:          databaseUrl,
		SignUpEmailWhitelist: signUpEmailWhitelist,
		SessionCookieSecure:  sessionCookieSecure,
		SessionSecretKey:     sessionSecretKey,
	}
}

func getStringEnvironmentVariable(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getBooleanEnvironmentVariable(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return strings.EqualFold(value, "true")
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
