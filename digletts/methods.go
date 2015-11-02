package digletts

import (
	"fmt"
)

type MethodName string

const (
	GetTile MethodName = "get_tile"
)

type Param string

const (
	Tileset Param = "tileset"
)

type MethodFunc func(params map[string]interface{}) (msg *ResponseMessage, err error)

type Methods struct {
	Methods map[MethodName]MethodFunc
}

var methods = &Methods{
	Methods: map[MethodName]MethodFunc{GetTile: getTileFunc},
}

func (m *Methods) Execute(method MethodName, params map[string]interface{}) (*ResponseMessage, error) {
	if f, ok := m.Methods[method]; ok {
		return f(params)
	}
	return nil, fmt.Errorf("The method does not exist! %q", method)
}

func getTileFunc(params map[string]interface{}) (msg *ResponseMessage, err error) {
	err = validateParams(params, []string{"tileset", "x", "y", "z"})
	if err != nil {
		return
	}
	slug, err := assertString(params, "tileset")
	if err != nil {
		return
	}
	ts, ok := tilesets.Tilesets[slug]
	if !ok {
		err = fmt.Errorf("Cannot find tileset %q", slug)
		return
	}
	x, err := assertInt(params, "x")
	if err != nil {
		return
	}
	y, err := assertInt(params, "y")
	if err != nil {
		return
	}
	z, err := assertInt(params, "z")
	if err != nil {
		return
	}
	tile, err := ts.ReadSlippyTile(x, y, z)
	msg = SuccessMsg(tile)
	return
}
