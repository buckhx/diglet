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

var methods = MethodIndex{Methods: map[string]Method{
	GetTile: Method{
		Name: GetTile,
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to read from"},
			"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
			"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
			"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
		},
		Handler: func(params MethodParams) (tile interface{}, err *CodedError) {
			x := params["x"].GetInt()
			y := params["y"].GetInt()
			z := params["z"].GetInt()
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; !ok {
				err = cerrorf(RpcInvalidRequest, "Cannot find tileset %s", slug)
			} else {
				var tserr error
				if tile, tserr = ts.ReadSlippyTile(x, y, z); tserr != nil {
					err = cerrorf(RpcInvalidRequest, tserr.Error())
				}
			}
			return
		},
		Help: "Retrieve a tile, the response's data field will be binary of the contents",
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
		Help: "List all of the tilesets available, including their metadata",
	},
	GetTileset: Method{
		Name:  GetTileset,
		Route: "/{tileset}",
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to query for metadata"},
		},
		Handler: func(params MethodParams) (attrs interface{}, err *CodedError) {
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; ok {
				attrs = ts.Metadata().Attributes()
			} else {
				err = cerrorf(RpcInvalidRequest, "No tileset named %s", slug)
			}
			return
		},
		Help: "Query for the tilesets metadata, all values are string representations",
	},
}}
