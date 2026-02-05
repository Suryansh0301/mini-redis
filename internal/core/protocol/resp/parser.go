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

func getParseNeedMoreDataResp() ParseResp {
	return ParseResp{
		result:        constants.ResultTypeNeedMoreData,
		value:         "",
		bytesConsumed: 0,
		err:           nil,
	}
}

func getParseErrorResp(err error) ParseResp {
	return ParseResp{
		result:        constants.ResultTypeError,
		value:         "",
		bytesConsumed: 0,
		err:           err,
	}
}

/*
TEST CASES:-
\r\n
+OK\r\n
?\r\n
+OK\r\n+PONG\r\n
$3\r\nfoo\r\n
:-\r\n
:00001\r\n
$-1\r\n -> null bulk string a valid case
$0\r\n\r\n -> empty bulk string also a valid case
*/

func Parse(buffer []byte) ParseResp {
	if len(buffer) == 0 {
		return getParseNeedMoreDataResp()
	}

	typeByte := buffer[0]
	index := readLine(buffer[1:])
	if index == -1 {
		return getParseNeedMoreDataResp()
	}

	return checkBuffer(typeByte, index,buffer)

}

func checkBuffer(typeByte byte,index int, bufferValue []byte) ParseResp {
	switch typeByte {
	case '+':
		for i := range bufferValue[:index] {
			if bufferValue[i] == '\r' || bufferValue[i] == '\n' {
				return getParseErrorResp(fmt.Errorf("invalid value received in simple string : %q", bufferValue))
			}
		}

		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[index]),
			bytesConsumed: 1 + len(bufferValue),
			err:           nil,
		}
	case '-':
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[:index]),
			bytesConsumed: 1 + len(bufferValue),
			err:           nil,
		}
	case ':':
		if len(bufferValue[:index]) == 0 {
			return getParseErrorResp(fmt.Errorf("empty value received in simple integer"))
		}

		if len(bufferValue[:index]) > 1 && bufferValue[0] == '0' {
			return getParseErrorResp(fmt.Errorf("invalid value received in simple integer : %q", bufferValue))
		}

		for i := range bufferValue[:index] {
			if i == 0 && (bufferValue[i] == '-' || bufferValue[i] == '+') {
				if len(bufferValue) == 1 {
					return getParseErrorResp(fmt.Errorf("invalid value received in simple string : %q", bufferValue))
				}
				continue
			}

			if bufferValue[i] < '0' || bufferValue[i] > '9' {
				return getParseErrorResp(fmt.Errorf("invalid value received in simple integer : %q", bufferValue)))
			}
		}
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[:index]),
			bytesConsumed: 1 + len(bufferValue),
			err:           nil,
		}
		case '$':
			if len(bufferValue) == 0 {
				return getParseErrorResp(fmt.Errorf("empty value received in bulk string"))
			}

			if len(bufferValue[:index]) == 2 && string(bufferValue[:index]) == "-1" {
				return ParseResp{
					result:        constants.ResultTypeSuccess,
					value:         "", //in my opinion it should be empty because it is a null string
					bytesConsumed: 1 + len(bufferValue),
					err:           nil,
				}
			}

			if bufferValue[0] == '0' {
				if len(bufferValue[:index]) == 1 {
					return ParseResp{
						result:        constants.ResultTypeSuccess,
						value:         "", //in my opinion it should be empty because it is a null string
						bytesConsumed: 1 + len(bufferValue) ,
						err:           nil,
					}
				}

				return getParseErrorResp(fmt.Errorf("invalid value received in bulk string : %q", bufferValue))
			}

			length:=0
			for i := range bufferValue[:index] {
				if bufferValue[i] < '1' || bufferValue[i] > '9' {
					return getParseErrorResp(fmt.Errorf("invalid value received in bulk string : %q", bufferValue))
				}
				length=length*10+int(bufferValue[i])
			}

			if length+2 > len(bufferValue[index+2:]) {
				return getParseNeedMoreDataResp()
			}

			return ParseResp{
				result:        constants.ResultTypeSuccess,
				value:         string(bufferValue[index+2 : length]),
				bytesConsumed: 1 + len(bufferValue) ,
				err:           nil,
			}



	default:
		return getParseErrorResp(fmt.Errorf("invalid type received: %q", typeByte))
	}
}

func readLine(buffer []byte) (index int) {
	for i := range buffer {
		if buffer[i] == '\r' {
			if i+1 < len(buffer) && buffer[i+1] == '\n' {
				return i
			}
		}
	}

	return -1
}
