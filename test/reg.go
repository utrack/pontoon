package test

import "github.com/utrack/pontoon/sdesc"

func (h Handler) RegisterHTTP(mux sdesc.HTTPRouter) {
	mux.MethodFunc("GET", "/v1/products/iterate", h.iterateProducts)
}
