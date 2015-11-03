package digletts

const (
	GetTile string = "get_tile"
)

type Param string

const (
	Tileset Param = "tileset"
)

type MethodFunc func(params map[string]interface{}) (msg *ResponseMessage, rerr *ResponseError)

type Methods struct {
	Methods map[string]MethodFunc
}

var methods = &Methods{
	Methods: map[string]MethodFunc{GetTile: getTileFunc},
}

func (m *Methods) Execute(method string, params map[string]interface{}) (msg *ResponseMessage, rerr *ResponseError) {
	if f, ok := m.Methods[method]; ok {
		msg, rerr = f(params)
	} else {
		rerr = Errorm(RpcMethodNotFound, sprintf("The method does not exist! %q", method))
	}
	return
}

func getTileFunc(params map[string]interface{}) (msg *ResponseMessage, rerr *ResponseError) {
	err := validateParams(params, []string{"tileset", "x", "y", "z"})
	if err != nil {
		rerr = Errorm(RpcInvalidParams, err.Error())
		return
	}
	slug, err := assertString(params, "tileset")
	if err != nil {
		rerr = Errorm(RpcParseError, err.Error())
		return
	}
	ts, ok := tilesets.Tilesets[slug]
	if !ok {
		rerr = Errorm(RpcInvalidRequest, sprintf("Cannot find tileset %q", slug))
		return
	}
	x, err := assertInt(params, "x")
	if err != nil {
		rerr = Errorm(RpcParseError, err.Error())
		return
	}
	y, err := assertInt(params, "y")
	if err != nil {
		rerr = Errorm(RpcParseError, err.Error())
		return
	}
	z, err := assertInt(params, "z")
	if err != nil {
		rerr = Errorm(RpcParseError, err.Error())
		return
	}
	tile, err := ts.ReadSlippyTile(x, y, z)
	if err != nil {
		rerr = Errorm(RpcInvalidRequest, err.Error())
	} else {
		msg = SuccessMsg(tile)
	}
	return
}
