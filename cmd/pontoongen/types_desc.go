package main

import "go/types"

var typeCache = map[types.Type]*typeDesc{}

type specialTypeType = uint

const (
	specialTypeNone specialTypeType = iota
	specialTypeTime
)

type typeDesc struct {
	id       string
	name     string
	doc      string
	isScalar bool

	isSpecial specialTypeType
	isStruct  *descStruct
	isSlice   *descSlice
	isMap     *descMap
	isPtr     *typeDesc
}

type descSlice struct {
	t *typeDesc
}

type descMap struct {
	key   *typeDesc
	value *typeDesc
}

type descStruct struct {
	embeds []descField
	fields []descField
}

type descField struct {
	name string
	doc  string
	tags string
	t    *typeDesc
}
