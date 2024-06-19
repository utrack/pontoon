package main

type serviceDesc struct {
	name string
	doc  string

	filename          string
	pkg               string
	serviceStructName string

	handlers []hdlDesc
}

type hdlDesc struct {
	goFuncName  string
	httpVerb    string
	path        string
	description string
	inout       hdlTypesDesc
}

type hdlTypesDesc struct {
	inType            *typeDesc
	hasResponseWriter bool
	outType           *typeDesc
	description       string
}
