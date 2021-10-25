package sdesc

import (
	"net/http"
)

// HandlerConfig configures Handler that passes RPC calls through
// to the Service.
type HandlerConfig struct {
	middlewares []func(http.Handler) http.Handler
}

func (c HandlerConfig) Clone() HandlerConfig {
	ret := HandlerConfig{}
	ret.middlewares = append(ret.middlewares, c.middlewares...)
	return ret
}

// Middlewares returns a collection of middlewares that
// should be applied to a single Service.
func (c HandlerConfig) Middlewares() []func(http.Handler) http.Handler {
	return c.middlewares
}

type ServiceOption func(*HandlerConfig)

// WithMiddlewares appends given middlewares that would be applied
// for a single Service.
func WithMiddlewares(mws ...func(http.Handler) http.Handler) ServiceOption {
	return func(c *HandlerConfig) {
		c.middlewares = append(c.middlewares, mws...)
	}
}
