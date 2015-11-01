package digletts

import "encoding/json"

const (
	RpcParseError          int = -32700
	RpcInvalidRequestError int = -32600
	RpcMethodNotFoundError int = -32601
	RpcInvalidParamsError  int = -32602
	RpcInternalError       int = -32603
	RpcServerError         int = -32000
)

type RequestMessage struct {
	Id      *uint                  `json:"id"`
	JsonRpc string                 `json:"jsonrpc"`
	Method  *string                `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

func (msg *RequestMessage) Validate() (rErr *ResponseError) {
	if msg.JsonRpc != "2.0" {
		rErr = &ResponseError{Code: RpcInvalidRequestError, Message: "jsonrpc != 2.0"}
	}
	//TODO validate methods
	return
}

func LoadRequestMessage(data []byte) (msg *RequestMessage, rErr *ResponseError) {
	err := json.Unmarshal(data, &msg)
	if err != nil {
		rErr = &ResponseError{Code: RpcInvalidRequestError, Message: "JSON-RPC requires valid json with fields: {'id', 'jsonrpc', 'method', 'params'}"}
	} else {
		rErr = msg.Validate()
	}
	if rErr != nil {
		msg = nil
	}
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

func SuccessMsg(content interface{}) (msg *ResponseMessage) {
	msg = &ResponseMessage{
		Error:   nil,
		Id:      nil,
		JsonRpc: "2.0",
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
		JsonRpc: "2.0",
		Result:  nil,
	}
	return
}
