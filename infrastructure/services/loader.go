package services

import (
	"fmt"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/config"
)

type loaderMap map[string]func(secrets config.Config) (domain.Service, error)

// MapServiceLoader loads actual instances of services.
type MapServiceLoader struct {
	services     loaderMap
	configLoader config.Loader
}

// ForName returns service instance for service name.
func (m MapServiceLoader) ForName(serviceName string) (service domain.Service, err error) {
	loader, exists := m.services[serviceName]

	if !exists {
		return nil, fmt.Errorf("unknown service %q", serviceName)
	}

	secrets, err := m.configLoader.Load("secrets")
	if err != nil {
		return nil, err
	}

	return loader(secrets)
}
