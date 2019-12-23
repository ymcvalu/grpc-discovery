package registry

import (
	"errors"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
)

var ErrDupRegister = errors.New("duplicate register")
var ErrRegistryClosed = errors.New("has closed")

type Registry interface {
	Register(inst instance.Instance) <-chan error
	Close()
}
