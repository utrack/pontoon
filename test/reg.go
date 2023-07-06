package test

import (
	"net/http"

	"github.com/utrack/pontoon/sdesc"
)

func (h Handler) RegisterHTTP(mux sdesc.HTTPRouter) {
	// POST
	mux.MethodFunc(http.MethodPost, "/v1/products/iterate/create", h.iterateProducts)
	// GET
	mux.MethodFunc(http.MethodGet, "/v1/products/iterate", h.iterateProducts)
	// Same path as above, different ops
	mux.MethodFunc(http.MethodPost, "/v1/products/iterate", h.iterateProducts)
	// Different return types
	mux.MethodFunc(http.MethodGet, "/v1/test/return/return-nothing", h.zeroReturn)
	mux.MethodFunc(http.MethodGet, "/v1/test/return/interface", h.ifaceReturn)
	mux.MethodFunc(http.MethodGet, "/v1/test/return/interface-any", h.ifaceReturnAny)
	mux.MethodFunc(http.MethodGet, "/v1/test/return/slice", h.sliceReturn)
	mux.MethodFunc(http.MethodGet, "/v1/test/return/slice-in-struct", h.sliceInObjReturn)
	mux.MethodFunc(http.MethodGet, "/v1/test/return/map", h.mapReturn)

	mux.MethodFunc(http.MethodGet, "/v1/test/request/jsonWithDirective", h.jsonWithDirs)
}
