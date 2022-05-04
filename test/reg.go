package test

import (
	"net/http"

	"github.com/utrack/pontoon/sdesc"
)

func (h Handler) RegisterHTTP(mux sdesc.HTTPRouter) {
	mux.MethodFunc(http.MethodGet, "/v1/products/iterate", h.iterateProducts)
}
