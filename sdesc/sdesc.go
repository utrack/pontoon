package sdesc

// Service is a collection of endpoints.
type Service interface {
	ServiceOptions() []ServiceOption
	RegisterHTTP(HTTPRouter)
}

// RPCHandler is a function of type
// func(*http.Request,<input type>) (<output type>,error)
// or
// func(*http.Request) (<out>,error)
// waiting for generics /shrug
type RPCHandler interface{}

// Router routes HTTP requests around.
type HTTPRouter interface {
	MethodFunc(method, pattern string, hdl RPCHandler)
}
