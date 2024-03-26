package test

import (
	"net/http"

	"github.com/ggicci/httpin"
	"github.com/pkg/errors"
	"github.com/utrack/pontoon/sdesc"
	"github.com/utrack/pontoon/test2"
)

// Handler struct comment
type Handler struct {
}

// Request comment
// Request line 2
type iterateRequest struct {
	iterateEmbedded

	Foo int64 `in:"query=foo"`

	// Local describes Some Stuff(tm). Required field.
	Local string `in:"query=local;required"`

	LocalWithDefault string `in:"query=local_default;required;default=foobarbaz"`

	SliceStrings []string `foo:"bar"`

	Maps map[string]mapped `faa:"faw"`

	Recursive *iterateRequest `in:"body=json"`
}

type mapped struct{}

// iterateEmbedded has an Embed comment
type iterateEmbedded struct {
	// PageToken is a token for the next page.
	// Pass an empty PageToken if you want to request the first page,
	// and use a token from the response as PageToken to get the next page.
	PageToken string `in:"query=page_token"`
}

// jsonWithDirectives describes a JSON-marshaled request with additional 'in' directives.
type jsonWithDirectives struct {
	httpin.JSONBody

	RequiredWithDefault string `json:"with_default,omitempty" in:"required;default=1234"`

	RequiredOnly string `in:"required"`
}

type jsonWithArrayOfStructs struct {
	Ret []dummyStruct
}

type dummyStruct struct {
	DummyField string
}

var _ sdesc.Service = &Handler{}

// IterateProducts comment
// Includes imported package
func (h Handler) iterateProducts(r *http.Request, req iterateRequest) (*test2.IterateResponse, error) {
	return nil, errors.New("NIH")
}

func (h Handler) ifaceReturn(r *http.Request, req iterateRequest) (interface{}, error) {
	return nil, errors.New("NIH")
}

func (h Handler) ifaceReturnAny(r *http.Request, req iterateRequest) (any, error) {
	return nil, errors.New("NIH")
}

func (h Handler) zeroReturn(r *http.Request, req iterateRequest) error {
	return errors.New("NIH")
}

func (h Handler) sliceReturn(r *http.Request, req iterateRequest) ([]test2.IterateResponse, error) {
	return nil, errors.New("NIH")
}

func (h Handler) sliceInObjReturn(r *http.Request) (*jsonWithArrayOfStructs, error) {
	return nil, errors.New("NIH")
}

func (h Handler) mapReturn(r *http.Request, req iterateRequest) (map[string]test2.IterateResponse, error) {
	return nil, errors.New("NIH")
}

func (h Handler) jsonWithDirs(r *http.Request, req jsonWithDirectives) error {
	return errors.New("NIH")
}

func (h Handler) ServiceOptions() []sdesc.ServiceOption {
	return nil
}
