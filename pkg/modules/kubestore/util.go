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
	key = specialChars.ReplaceAllString(key, ".")
	return fmt.Sprintf(keyFormat, key)
}
