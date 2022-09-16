package kubestore

import (
	"fmt"
	"regexp"
)

const (
	keyPrefix = "tls-"
	keySuffix = "--acme"
	keyFormat = keyPrefix + "%s" + keySuffix
)

var specialChars = regexp.MustCompile(`[^\da-zA-Z-]+`)

func cleanKey(key string) string {
	return specialChars.ReplaceAllString(key, ".")
}

func generateSecretName(key string) string {
	return fmt.Sprintf(keyFormat, cleanKey(key))
}
