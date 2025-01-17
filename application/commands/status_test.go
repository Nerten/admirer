package commands

import (
	"bytes"
	"errors"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/dietrichm/admirer/domain"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	t.Run("returns status for each service", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		fooService := domain.NewMockService(ctrl)
		fooService.EXPECT().Name().Return("Foo")
		fooService.EXPECT().Authenticated().Return(true)
		fooService.EXPECT().GetUsername().Return("user303", nil)
		fooService.EXPECT().Close()

		barService := domain.NewMockService(ctrl)
		barService.EXPECT().Name().Return("Bar")
		barService.EXPECT().Authenticated().Return(true)
		barService.EXPECT().GetUsername().Return("user808", nil)
		barService.EXPECT().Close()

		serviceLoader := domain.NewMockServiceLoader(ctrl)
		serviceLoader.EXPECT().Names().Return([]string{"foo", "bar"})
		serviceLoader.EXPECT().ForName("foo").Return(fooService, nil)
		serviceLoader.EXPECT().ForName("bar").Return(barService, nil)

		expected := `Foo
	Authenticated as user303
Bar
	Authenticated as user808
`
		got, err := executeStatus(serviceLoader)

		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns error when failing to load service", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		expected := "failed to load"
		serviceLoader := domain.NewMockServiceLoader(ctrl)
		serviceLoader.EXPECT().Names().Return([]string{"foo"})
		serviceLoader.EXPECT().ForName("foo").Return(nil, errors.New(expected))

		output, err := executeStatus(serviceLoader)

		assert.EqualError(t, err, expected)
		assert.Empty(t, output)
	})

	t.Run("returns message when not authenticated", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		fooService := domain.NewMockService(ctrl)
		fooService.EXPECT().Name().Return("Foo")
		fooService.EXPECT().Authenticated().Return(false)
		fooService.EXPECT().Close()

		serviceLoader := domain.NewMockServiceLoader(ctrl)
		serviceLoader.EXPECT().Names().Return([]string{"foo"})
		serviceLoader.EXPECT().ForName("foo").Return(fooService, nil)

		expected := `Foo
	Not logged in
`
		got, err := executeStatus(serviceLoader)

		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns error when failed to get username", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		expected := "auth error"
		fooService := domain.NewMockService(ctrl)
		fooService.EXPECT().Authenticated().Return(true)
		fooService.EXPECT().GetUsername().Return("", errors.New(expected))
		fooService.EXPECT().Close()

		serviceLoader := domain.NewMockServiceLoader(ctrl)
		serviceLoader.EXPECT().Names().Return([]string{"foo"})
		serviceLoader.EXPECT().ForName("foo").Return(fooService, nil)

		output, err := executeStatus(serviceLoader)

		assert.EqualError(t, err, expected)
		assert.Empty(t, output)
	})
}

func executeStatus(serviceLoader domain.ServiceLoader) (string, error) {
	buffer := new(bytes.Buffer)
	err := status(serviceLoader, buffer)
	return buffer.String(), err
}
