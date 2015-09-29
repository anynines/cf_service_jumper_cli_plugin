package main

import (
	"errors"
	"strings"
)

func GetIdentityAndKey(sharedSecret string) (string, string, error) {
	sharedSecretCollection := strings.Split(sharedSecret, ":")

	if len(sharedSecretCollection) != 2 {
		return "", "", errors.New("Failed to get shared secret (PSK information).")
	}
	identity := sharedSecretCollection[0]
	key := sharedSecretCollection[1]

	if len(identity) < 1 {
		return "", "", errors.New("Failed to get shared secret (PSK information).")
	}
	if len(key) < 1 {
		return "", "", errors.New("Failed to get shared secret (PSK information).")
	}

	return identity, key, nil
}
