package digletts

const (
	GetTile      string = "get_tile"
	GetTileset   string = "get_tileset"
	ListTilesets string = "list_tilesets"
)

type Param struct {
	Key       string
	Value     interface{}
	Validator func(interface{}) error
}

func (p Param) GetInt() int {
	v := p.Value.(float64)
	return int(v)
}

func (p Param) GetString() string {
	return p.Value.(string)
}

func (p Param) Validate() error {
	return p.Validator(p.Value)
}

type MethodHandler func(params MethodParams) (interface{}, *CodedError)

type MethodParams map[string]Param

type Method struct {
	Name    string
	Handler MethodHandler
	Params  MethodParams
	Route   string
}

func (m Method) BuildParams(params map[string]interface{}) (mParams MethodParams, err error) {
	// TODO do we need a copy?
	mParams = make(MethodParams)
	for key, param := range m.Params {
		if raw, ok := params[key]; !ok {
			err = errorf("Missing param: %q", key)
		} else {
			if err = param.Validator(raw); err == nil {
				mParams[key] = Param{
					Key:       key,
					Value:     raw,
					Validator: param.Validator,
				}
			} else {
			}
		}
	}
	return
}

func (m Method) Execute(params MethodParams) (interface{}, *CodedError) {
	return m.Handler(params)
}

type MethodIndex struct {
	Methods map[string]Method
}

func (m *MethodIndex) Execute(methodName string, params map[string]interface{}) (msg *ResponseMessage) {
	var err *CodedError
	var content interface{}
	if method, ok := m.Methods[methodName]; !ok {
		err = cerrorf(RpcMethodNotFound, "The method does not exist! %q", method)
	} else {
		mParams, perr := method.BuildParams(params)
		if perr != nil {
			err = cerrorf(RpcInvalidParams, perr.Error())
		} else {
			content, err = method.Execute(mParams)
		}
	}
	if err != nil {
		msg = err.ResponseMessage()
	} else if content != nil {
		msg = SuccessMsg(content)
	}
	return
}

var methods = MethodIndex{Methods: map[string]Method{
	GetTile: Method{
		Name:  GetTile,
		Route: "/{tileset}/{z}/{x}/{y}",
		Params: MethodParams{
			"tileset": {Validator: assertString},
			"x":       {Validator: assertNumber},
			"y":       {Validator: assertNumber},
			"z":       {Validator: assertNumber},
		},
		Handler: func(params MethodParams) (tile interface{}, err *CodedError) {
			x := params["x"].GetInt()
			y := params["y"].GetInt()
			z := params["z"].GetInt()
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; !ok {
				err = cerrorf(RpcInvalidRequest, "Cannot find tileset %q", slug)
			} else {
				var tserr error
				if tile, tserr = ts.ReadSlippyTile(x, y, z); tserr != nil {
					err = cerrorf(RpcInvalidRequest, tserr.Error())
				}
			}
			return
		},
	},
	ListTilesets: Method{
		Name:   ListTilesets,
		Route:  "/",
		Params: MethodParams{},
		Handler: func(params MethodParams) (interface{}, *CodedError) {
			dict := make(map[string]map[string]string)
			for name, ts := range tilesets.Tilesets {
				dict[name] = ts.Metadata().Attributes()
			}
			return dict, nil
		},
	},
	GetTileset: Method{
		Name:  GetTileset,
		Route: "/",
		Params: MethodParams{
			"tileset": {Validator: assertString},
		},
		Handler: func(params MethodParams) (attrs interface{}, err *CodedError) {
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; ok {
				attrs = ts.Metadata().Attributes()
			} else {
				err = cerrorf(RpcInvalidRequest, "No tileset named %q", slug)
			}
			return
		},
	},
}}
