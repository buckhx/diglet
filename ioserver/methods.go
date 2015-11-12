package ioserver

type Param struct {
	Key       string                  `json:"key,omitempty"`
	Value     interface{}             `json:"value,omitempty"`
	Validator func(interface{}) error `json:"-"`
	Help      string                  `json:"help,omitempty"`
}

func (p *Param) GetInt() int {
	v := p.Value.(float64)
	return int(v)
}

func (p *Param) GetUint() uint {
	v := p.Value.(float64)
	return uint(v)
}

func (p *Param) GetString() string {
	return p.Value.(string)
}

func (p *Param) Validate() error {
	return p.Validator(p.Value)
}

type MethodHandler func(ctx *RequestContext) (interface{}, *CodedError)

type MethodParams map[string]*Param

type Method struct {
	Name    string        `json:"name"`
	Handler MethodHandler `json:"-"`
	Params  MethodParams  `json:"params,omitempty"`
	Route   string        `json:"route,omitempty"`
	Help    string        `json:"help,omitempty"`
}

func (m Method) WrapParams(params map[string]interface{}) (MethodParams, error) {
	// TODO do we need a copy?
	mParams := make(MethodParams)
	for key, param := range m.Params {
		if raw, ok := params[key]; !ok {
			return nil, errorf("Missing param: %s", key)
		} else {
			if err := param.Validator(raw); err == nil {
				mParams[key] = &Param{
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

func (m Method) Execute(ctx *RequestContext) (msg *ResponseMessage) {
	params, perr := m.WrapParams(ctx.Request.Params)
	if perr != nil {
		msg = cerrorf(RpcInvalidParams, perr.Error()).ResponseMessage()
	} else {
		ctx.Params = params
		res, err := m.Handler(ctx)
		if res != nil {
			msg = SuccessMsg(res)
		}
		if err != nil {
			msg = err.ResponseMessage()
		}
	}
	if msg != nil {
		msg.Id = ctx.Request.Id
	}
	return
}
