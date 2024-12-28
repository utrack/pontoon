package httpinoapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerator_Operation(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}

	cases := []struct {
		name    string
		verb    string
		path    string
		handler interface{}
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid_get",
			verb:    http.MethodGet,
			path:    "/users/:id",
			handler: handler,
			opts: []Option{
				WithTags("users"),
			},
			wantErr: false,
		},
		{
			name:    "valid_post_with_types",
			verb:    http.MethodPost,
			path:    "/users",
			handler: handler,
			opts: []Option{
				WithTags("users"),
				WithInputStruct(struct{ Name string }{}),
				WithOutputStruct(struct{ ID int }{}),
			},
			wantErr: false,
		},
		{
			name:    "invalid_verb",
			verb:    "INVALID",
			path:    "/users",
			handler: handler,
			wantErr: true,
		},
		{
			name:    "empty_path",
			verb:    http.MethodGet,
			path:    "",
			handler: handler,
			wantErr: true,
		},
		{
			name:    "invalid_path",
			verb:    http.MethodGet,
			path:    "users",
			handler: handler,
			wantErr: true,
		},
		{
			name:    "nil_handler",
			verb:    http.MethodGet,
			path:    "/users",
			handler: nil,
			wantErr: true,
		},
		{
			name:    "non_func_handler",
			verb:    http.MethodGet,
			path:    "/users",
			handler: "not a func",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGenerator()
			err := g.Operation(tc.verb, tc.path, tc.handler, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check that handler was registered
			require.Len(t, g.handlers, 1)
			h := g.handlers[0]
			require.Equal(t, tc.verb, h.verb)
			require.Equal(t, tc.path, h.path)
			require.NotNil(t, h.handler)
		})
	}
}

func TestGenerator_Operation_Duplicates(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// First registration should succeed
	err := g.Operation(http.MethodGet, "/users", handler)
	require.NoError(t, err)

	// Second registration with same verb+path should fail
	err = g.Operation(http.MethodGet, "/users", handler)
	require.Error(t, err)

	// Different verb should succeed
	err = g.Operation(http.MethodPost, "/users", handler)
	require.NoError(t, err)

	// Different path should succeed
	err = g.Operation(http.MethodGet, "/users/:id", handler)
	require.NoError(t, err)
}
