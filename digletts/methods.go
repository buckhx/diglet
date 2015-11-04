package digletts

const (
	GetTile string = "get_tile"
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

type MethodHandler func(params MethodParams) (*interface{}, *CodedError)

type MethodParams map[string]Param

type Method struct {
	Name    string
	Handler MethodHandler
	Params  MethodParams
	Route   string
}

func (m Method) GatherParams(params map[string]interface{}) (err error) {
	for key, param := range m.Params {
		if raw, ok := params[key]; !ok {
			err = errorf("Missing param: %q", key)
		} else {
			if err = param.Validator(raw); err == nil {
				// Probably needs to be a reference
				param.Key = key
				param.Value = raw
			}
		}
	}
	return
}

func (m Method) Execute() (*interface{}, *CodedError) {
	return m.Handler(m.Params)
}

type MethodIndex struct {
	Methods map[string]Method
}

func (m *MethodIndex) Execute(methodName string, params map[string]interface{}) (msg *ResponseMessage) {
	var err *CodedError
	var content *interface{}
	if method, ok := m.Methods[methodName]; !ok {
		err = cerrorf(RpcMethodNotFound, "The method does not exist! %q", method)
	} else {
		perr := method.GatherParams(params)
		if perr != nil {
			err = cerrorf(RpcInvalidParams, perr.Error())
		} else {
			content, err = method.Execute()
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
		Handler: func(params MethodParams) (tile *interface{}, err *CodedError) {
			x := params["x"].GetInt()
			y := params["y"].GetInt()
			z := params["z"].GetInt()
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; !ok {
				err = cerrorf(RpcInvalidRequest, "Cannot find tileset %q", slug)
			} else if tile, tserr := ts.ReadSlippyTile(x, y, z); tserr != nil {
				err = cerrorf(RpcInvalidRequest, tserr.Error())
			}
			return
		},
	},
}}
