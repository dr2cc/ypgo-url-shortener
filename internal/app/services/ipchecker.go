package services

import "net/http"

type IPCheckerInterface interface {
	IsRequestFromTrustedSubnet(r *http.Request) (bool, error)
}

type IPChecker struct{}

func (c *IPChecker) IsRequestFromTrustedSubnet(r *http.Request) (bool, error) {
	// TODO: implement
	return false, nil
}
