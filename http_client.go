package main

import (
	"crypto/tls"

	"github.com/parnurzeal/gorequest"
)

func NewHttpClient(isSSLDisabled bool) *gorequest.SuperAgent {
	request := gorequest.New()
	if isSSLDisabled {
		request = request.TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	return request
}
