package test

import (
	"net/http"

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
	// Local field
	Local string `in:"query=local"`

	SliceStrings []string `foo:"bar"`

	Maps map[string]mapped `faa:"faw"`

	Recursive *iterateRequest `in:"body=json"`
}

type mapped struct{}

// Embed comment
type iterateEmbedded struct {
	// PageToken comment
	PageToken string `in:"query=page_token"`
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

func (h Handler) mapReturn(r *http.Request, req iterateRequest) (map[string]test2.IterateResponse, error) {
	return nil, errors.New("NIH")
}

func (h Handler) ServiceOptions() []sdesc.ServiceOption {
	return nil
}
