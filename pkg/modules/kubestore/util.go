package kubestore

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	keyPrefix = "tls-"
	keySuffix = "--acme"
	keyFormat = keyPrefix + "%s" + keySuffix

	secretKeyPrefix = "secret/"
)

var specialChars = regexp.MustCompile(`[^\da-zA-Z-]+`)

func cleanKey(key string) string {
	return specialChars.ReplaceAllString(key, ".")
}

func generateSecretName(key string) string {
	if strings.HasPrefix(key, secretKeyPrefix) {
		return strings.Split(key, "/")[2]
	}
	return fmt.Sprintf(keyFormat, cleanKey(key))
}

func getNamespace(key string, def string) string {
	if strings.HasPrefix(key, secretKeyPrefix) {
		return strings.Split(key, "/")[1]
	}
	return def
}

func getDataKey(key string) string {
	if strings.HasPrefix(key, secretKeyPrefix) {
		return strings.Split(key, "/")[3]
	}
	return dataKey
}
