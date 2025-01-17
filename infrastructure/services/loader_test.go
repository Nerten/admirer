package services

import (
	"errors"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/config"
)

func TestMapServiceLoader(t *testing.T) {
	t.Run("returns service when loader exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		service := domain.NewMockService(ctrl)
		secrets := config.NewMockConfig(ctrl)

		configLoader := config.NewMockLoader(ctrl)
		configLoader.EXPECT().Load("secrets-foo").Return(secrets, nil)

		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return service, nil
				},
				"bar": func(secrets config.Config) (domain.Service, error) {
					return nil, nil
				},
			},
			configLoader: configLoader,
		}

		got, err := serviceLoader.ForName("foo")

		if got != service {
			t.Errorf("expected %v, got %v", service, got)
		}

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("returns service for name with mixed capitals and punctuation", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		service := domain.NewMockService(ctrl)
		secrets := config.NewMockConfig(ctrl)

		configLoader := config.NewMockLoader(ctrl)
		configLoader.EXPECT().Load(gomock.Any()).Return(secrets, nil).AnyTimes()

		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return service, nil
				},
			},
			configLoader: configLoader,
		}

		for _, name := range []string{"Foo", "FOO", "Fo.o", "Fo-o"} {
			got, err := serviceLoader.ForName(name)

			if got != service {
				t.Errorf("expected %v, got %v", service, got)
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	})

	t.Run("returns error when loader does not exist", func(t *testing.T) {
		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return nil, nil
				},
			},
		}

		service, err := serviceLoader.ForName("bar")

		if service != nil {
			t.Errorf("Unexpected service instance: %v", service)
		}

		if err == nil {
			t.Fatal("Expected an error")
		}

		expected := `unknown service "bar"`
		got := err.Error()

		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns error when configuration fails to load", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		configLoader := config.NewMockLoader(ctrl)
		configError := errors.New("failed to load")
		configLoader.EXPECT().Load(gomock.Any()).Return(nil, configError)

		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return nil, nil
				},
			},
			configLoader: configLoader,
		}

		service, err := serviceLoader.ForName("foo")

		if service != nil {
			t.Errorf("Unexpected service instance: %v", service)
		}

		if err == nil {
			t.Fatal("Expected an error")
		}

		if err != configError {
			t.Errorf("expected %q, got %q", configError, err)
		}
	})

	t.Run("returns error when loader yields error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		secrets := config.NewMockConfig(ctrl)
		configLoader := config.NewMockLoader(ctrl)
		configLoader.EXPECT().Load(gomock.Any()).Return(secrets, nil)

		serviceError := errors.New("service error")
		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return nil, serviceError
				},
			},
			configLoader: configLoader,
		}

		service, err := serviceLoader.ForName("foo")

		if service != nil {
			t.Errorf("Unexpected service instance: %v", service)
		}

		if err == nil {
			t.Fatal("Expected an error")
		}

		if err != serviceError {
			t.Errorf("expected %q, got %q", serviceError, err)
		}
	})

	t.Run("returns slice of names of available services", func(t *testing.T) {
		serviceLoader := mapServiceLoader{
			services: loaderMap{
				"foo": func(secrets config.Config) (domain.Service, error) {
					return nil, nil
				},
				"bar": func(secrets config.Config) (domain.Service, error) {
					return nil, nil
				},
			},
		}

		expected := []string{"bar", "foo"}
		got := serviceLoader.Names()

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
