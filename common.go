package jsonrpc

type (
	Request struct {
		JsonRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		Id      uint64      `json:"id"`
	}

	Response struct {
		JsonRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		Error   interface{} `json:"error"`
		Id      uint64      `json:"id"`
	}
)
