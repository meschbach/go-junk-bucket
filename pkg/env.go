package pkg

import "os"

// EnvOrDefault will use the value of keyName if defined in the environment, otherwise the value defaultValue is returned.
func EnvOrDefault(keyName, defaultValue string) string {
	if value, ok := os.LookupEnv(keyName); ok {
		return value
	} else {
		return defaultValue
	}
}
