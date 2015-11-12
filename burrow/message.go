package burrow

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
	Id      *string                `json:"id"`
	JsonRpc string                 `json:"jsonrpc"`
	Method  *string                `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

func (req *RequestMessage) Validate() (err *CodedError) {
	switch {
	case req.JsonRpc != RpcVersion:
		err = cerrorf(RpcInvalidRequest, "jsonrpc != "+RpcVersion)
	case req.Method == nil:
		err = cerrorf(RpcInvalidRequest, "Request is missing field 'method'")
	case req.Params == nil:
		// still want to inject params even if they weren't passed
		req.Params = make(map[string]interface{})
	}
	return
}

func (req *RequestMessage) String() string {
	if b, err := json.Marshal(req); err != nil {
		warn(err, "Could not marshal tile_xyz")
		return sprintf("Could not marshal tile_xyz %s", req)
	} else {
		return string(b)
	}
}

func (req *RequestMessage) MethodName() string {
	return *req.Method
}

func LoadRequestMessage(data []byte) (msg *RequestMessage, err *CodedError) {
	if merr := json.Unmarshal(data, &msg); merr != nil {
		hint := "JSON-RPC requires valid json with fields: {'id', 'jsonrpc', 'method','params'}"
		err = cerrorf(RpcInvalidRequest, hint)
	} else {
		err = msg.Validate()
	}
	return
}

func ReadRequestMessage(content io.Reader) (msg *RequestMessage, err *CodedError) {
	if body, ioerr := ioutil.ReadAll(content); ioerr != nil {
		err = cerrorf(RpcParseError, "Could not read body")
	} else {
		msg, err = LoadRequestMessage(body)
	}
	return
}

type ResponseMessage struct {
	Error   *CodedError `json:"error,omitempty"`
	Id      *string     `json:"id,omitempty"`
	JsonRpc string      `json:"jsonrpc,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func (msg *ResponseMessage) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type CodedError struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

func (err *CodedError) Error() string {
	return sprintf("Error %s: %s", err.Code, err.Message)
}

func (err *CodedError) ResponseMessage() (msg *ResponseMessage) {
	// not sure if Error: should be refernce to this
	msg = &ResponseMessage{
		Error: &CodedError{
			Code:    err.Code,
			Data:    err.Data,
			Message: err.Message,
		},
		Id:      nil,
		JsonRpc: RpcVersion,
		Result:  nil,
	}
	return
}

//TODO too lazy to actually rename this right now
func cerrorf(code int, msg string, vals ...interface{}) (err *CodedError) {
	err = &CodedError{
		Code:    code,
		Message: sprintf(msg, vals...),
	}
	return
}

func Cerrorf(code int, msg string, vals ...interface{}) (err *CodedError) {
	err = &CodedError{
		Code:    code,
		Message: sprintf(msg, vals...),
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
