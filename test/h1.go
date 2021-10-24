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
	Local string `in:"body=json"`

	SliceStrings []string `foo:"bar"`

	Maps map[string]mapped `faa:"faw"`
}

type mapped struct{}

// Embed comment
type iterateEmbedded struct {
	// PageToken comment
	PageToken string `in:"query=page_token"`
}

var _ sdesc.Service = &Handler{}

// IterateProducts comment
func (h Handler) iterateProducts(r *http.Request, req iterateRequest) (*test2.IterateResponse, error) {
	return nil, errors.New("NIH")
}

func (h Handler) ServiceOptions() []sdesc.ServiceOption {
	return nil
}
