package main

import (
	"fmt"
	"strconv"
)

type ForwardSbCredentials map[string]interface{}

func (self ForwardSbCredentials) CredentialsMap() map[string]string {
	credentials := make(map[string]string)

	for k, v := range self {
		switch value := v.(type) {
		case string:
			credentials[k] = value
		case int:
			credentials[k] = strconv.Itoa(value)
		case float64:
			credentials[k] = fmt.Sprint(value)
		default:
		}
	}
	return credentials
}

type ForwardCredentials struct {
	Credentials ForwardSbCredentials `json:"credentials"`
}

type ForwardDataSet struct {
	ID           int                `json:"id"`
	Hosts        []string           `json:"public_uris"`
	SharedSecret string             `json:"shared_secret"`
	Credentials  ForwardCredentials `json:"credentials"`
}

// Returns map with credential information
func (self ForwardDataSet) CredentialsMap() map[string]string {
	return self.Credentials.Credentials.CredentialsMap()
}
