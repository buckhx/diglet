package digletts

const (
	GetTile      string = "get_tile"
	GetTileset   string = "get_tileset"
	ListTilesets string = "list_tilesets"
)

type Param struct {
	Key       string                  `json:"key,omitempty"`
	Value     interface{}             `json:"value,omitempty"`
	Validator func(interface{}) error `json:"-"`
	Help      string                  `json:"help,omitempty"`
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
	Name    string        `json:"name"`
	Handler MethodHandler `json:"-"`
	Params  MethodParams  `json:"params,omitempty"`
	Route   string        `json:"route,omitempty"`
	Help    string        `json:"help,omitempty"`
}

func (m Method) BuildParams(params map[string]interface{}) (MethodParams, error) {
	// TODO do we need a copy?
	mParams := make(MethodParams)
	for key, param := range m.Params {
		if raw, ok := params[key]; !ok {
			return nil, errorf("Missing param: %s", key)
		} else {
			if err := param.Validator(raw); err == nil {
				mParams[key] = Param{
					Key:       key,
					Value:     raw,
					Validator: param.Validator,
				}
			} else {
				return nil, errorf("Invalid param %s: %s", key, raw)
			}
		}
	}
	return mParams, nil
}

func (m Method) Execute(params MethodParams) (interface{}, *CodedError) {
	return m.Handler(params)
}

// Route order is not guaranteed here, might want to have a list instead w/ map view
type MethodIndex struct {
	Methods map[string]Method
}

func (m *MethodIndex) Execute(methodName string, params map[string]interface{}) (msg *ResponseMessage) {
	var err *CodedError
	var content interface{}
	if method, ok := m.Methods[methodName]; !ok {
		err = cerrorf(RpcMethodNotFound, "The method does not exist! %s", method)
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
