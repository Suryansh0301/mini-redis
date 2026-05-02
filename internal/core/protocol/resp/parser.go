package resp

import (
	"fmt"

	"github.com/suryansh0301/mini-redis/internal/constants"
)

type ParseResp struct {
	result        constants.ResultType
	value         string
	bytesConsumed int
	err           error
}

func Parse(buffer []byte) ParseResp {
	if buffer[0] == '+' || buffer[0] == '-' || buffer[0] == ':' || buffer[0] == '*' || buffer[0] == '$' {
		// added need more data as asked
		return parseResp{
			result:        constants.ResultTypeNeedMoreData,
			value:         "",
			bytesConsumed: 0,
			err:           nil,
		}
	} else { // return error if first character of buffer is invalid
		return parseResp{
			result:        constants.ResultTypeError,
			value:         "",
			bytesConsumed: 0,
			err:           fmt.Errorf("Invalid type received:" + string(buffer[0])),
		}
	}
}
