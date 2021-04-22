//go:generate mockgen -source authentication.go -destination authentication_mock.go -package authentication

package authentication

// DefaultCallbackProvider is the default callback provider for services.
var DefaultCallbackProvider = &httpCallbackProvider{}

// CallbackProvider provides a callback mechanism for authenticating services.
type CallbackProvider interface{}