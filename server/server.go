package server

import (
	"context"
)

type Server interface {
	ListenAndServe(context.Context) error
}
