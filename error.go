package main

import (
	"github.com/6thplaneta/hermes"
)

func newError(key, desc string) hermes.Error {
	return hermes.Error{Key: key, Decsription: desc}
}

var ErrNoPayment = newError("NoPayment", "Invalid payment token!")
var ErrReported = newError("Reported", "User is reported!")
var ErrBlocked = newError("Blocked", "User is blocked!")
var ErrConfigNotFound = newError("ConfigNotFound", "Config not found!")
var ErrStatusError = newError("StatusError", "Status is incorrect!")
var ErrElasticError = newError("ElasticError", "Elasticsearch connecting error!")
