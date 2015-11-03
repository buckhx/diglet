package digletts

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

const (
	RpcVersion        string = "2.0"
	RpcParseError     int    = -32700
	RpcInvalidRequest int    = -32600
	RpcMethodNotFound int    = -32601
	RpcInvalidParams  int    = -32602
	RpcInternalError  int    = -32603
	RpcServerError    int    = -32000
)

type RequestMessage struct {
	Id      *uint                  `json:"id"`
	JsonRpc string                 `json:"jsonrpc"`
	Method  *string                `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

func (req *RequestMessage) Validate() (rerr *ResponseError) {
	if req.JsonRpc != RpcVersion {
		rerr = Errorm(RpcInvalidRequest, "jsonrpc != "+RpcVersion)
	}
	if req.Method == nil {
		rerr = Errorm(RpcInvalidRequest, "Request is missing field 'method'")
	}
	return
}

func (req *RequestMessage) ExecuteMethod() (msg *ResponseMessage) {
	msg, rerr := methods.Execute(*req.Method, req.Params)
	if rerr != nil {
		msg = ErrorMsg(rerr.Code, rerr.Message)
	}
	msg.Id = req.Id
	return
}

func LoadRequestMessage(data []byte) (msg *RequestMessage, rerr *ResponseError) {
	err := json.Unmarshal(data, &msg)
	if err != nil {
		rerr = Errorm(RpcInvalidRequest, "JSON-RPC requires valid json with fields: {'id', 'jsonrpc', 'method', 'params'}")
	} else {
		rerr = msg.Validate()
	}
	if rerr != nil {
		//TODO check if this is neceaary
		msg = nil
	}
	return
}

func ReadRequestMessage(content io.Reader) (msg *RequestMessage, rerr *ResponseError) {
	body, err := ioutil.ReadAll(content)
	if err != nil {
		rerr = &ResponseError{Code: RpcParseError, Message: "Could not read body"}
	}
	msg, rerr = LoadRequestMessage(body)
	return
}

type ResponseMessage struct {
	Error   *ResponseError `json:"error"`
	Id      *uint          `json:"id"`
	JsonRpc string         `json:"jsonrpc"`
	Result  interface{}    `json:"result"`
}

func (msg *ResponseMessage) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type ResponseError struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Errorm(code int, message string) (rerr *ResponseError) {
	rerr = &ResponseError{
		Code:    code,
		Message: message,
	}
	return
}

func SuccessMsg(content interface{}) (msg *ResponseMessage) {
	msg = &ResponseMessage{
		Error:   nil,
		Id:      nil,
		JsonRpc: RpcVersion,
		Result:  content,
	}
	return
}

func ErrorMsg(code int, message string) (msg *ResponseMessage) {
	msg = &ResponseMessage{
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
		Id:      nil,
		JsonRpc: RpcVersion,
		Result:  nil,
	}
	return
}
