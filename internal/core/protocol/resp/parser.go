package resp

import (
	"strconv"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

type ParseResp struct {
	statusCode    enums.StatusCode
	resp          *common.RespValue
	bytesConsumed int
	err           error
}

func getParseNeedMoreDataResp() ParseResp {
	return ParseResp{
		statusCode:    enums.NeedMoreDataStatusCode,
		bytesConsumed: 0,
	}
}

func getParseErrorResp(err error) ParseResp {
	return ParseResp{
		statusCode:    enums.ErrorStatusCode,
		bytesConsumed: 0,
		err:           err,
	}
}

func Parse(buffer []byte) ParseResp {
	if len(buffer) == 0 {
		return getParseNeedMoreDataResp()
	}

	typeByte := buffer[0]
	index := readLine(buffer[1:])
	if index == -1 {
		return getParseNeedMoreDataResp()
	}

	return checkBuffer(typeByte, index, buffer[1:])
}

func checkBuffer(typeByte byte, index int, bufferValue []byte) ParseResp {
	switch typeByte {

	case '+':
		for i := 0; i < index; i++ {
			if bufferValue[i] == '\r' || bufferValue[i] == '\n' {
				return getParseErrorResp(common.ProtocolError("invalid simple string"))
			}
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &common.RespValue{
				Type: enums.SimpleStringRespType,
				Str:  string(bufferValue[:index]),
			},
			bytesConsumed: 1 + index + 2,
		}

	case '-':
		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &common.RespValue{
				Type: enums.ErrorRespType,
				Str:  string(bufferValue[:index]),
			},
			bytesConsumed: 1 + index + 2,
		}

	case ':':
		if index == 0 {
			return getParseErrorResp(common.ProtocolError("invalid integer"))
		}

		val, err := strconv.ParseInt(string(bufferValue[:index]), 10, 64)
		if err != nil {
			return getParseErrorResp(common.ProtocolError("invalid integer"))
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &common.RespValue{
				Type: enums.IntRespType,
				Int:  val,
			},
			bytesConsumed: 1 + index + 2,
		}

	case '$':
		if index == 0 {
			return getParseErrorResp(common.ProtocolError("invalid bulk length"))
		}

		length64, err := strconv.ParseInt(string(bufferValue[:index]), 10, 64)
		if err != nil {
			return getParseErrorResp(common.ProtocolError("invalid bulk length"))
		}

		// Null bulk string
		if length64 == -1 {
			return ParseResp{
				statusCode: enums.SuccessStatusCode,
				resp: &common.RespValue{
					Type:   enums.BulkStringRespType,
					IsNull: true,
				},
				bytesConsumed: 1 + index + 2,
			}
		}

		if length64 < 0 {
			return getParseErrorResp(common.ProtocolError("invalid bulk length"))
		}

		length := int(length64)
		payloadStart := index + 2
		required := payloadStart + length + 2

		if len(bufferValue) < required {
			return getParseNeedMoreDataResp()
		}

		if bufferValue[payloadStart+length] != '\r' ||
			bufferValue[payloadStart+length+1] != '\n' {
			return getParseErrorResp(common.ProtocolError("invalid bulk string"))
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &common.RespValue{
				Type: enums.BulkStringRespType,
				Str:  string(bufferValue[payloadStart : payloadStart+length]),
			},
			bytesConsumed: 1 + index + 2 + length + 2,
		}

	case '*':
		if index == 0 {
			return getParseErrorResp(common.ProtocolError("invalid array length"))
		}

		length64, err := strconv.ParseInt(string(bufferValue[:index]), 10, 64)
		if err != nil {
			return getParseErrorResp(common.ProtocolError("invalid array length"))
		}

		if length64 == -1 {
			return ParseResp{
				statusCode: enums.SuccessStatusCode,
				resp: &common.RespValue{
					Type:   enums.ArrayRespType,
					IsNull: true,
				},
				bytesConsumed: 1 + index + 2,
			}
		}

		if length64 < 0 {
			return getParseErrorResp(common.ProtocolError("invalid array length"))
		}

		length := int(length64)

		totalConsumed := 1 + index + 2
		cursor := index + 2

		values := make([]*common.RespValue, 0, length)

		for i := 0; i < length; i++ {
			response := Parse(bufferValue[cursor:])

			if response.statusCode == enums.ErrorStatusCode {
				return getParseErrorResp(response.err)
			}

			if response.statusCode == enums.NeedMoreDataStatusCode {
				return getParseNeedMoreDataResp()
			}

			values = append(values, response.resp)
			cursor += response.bytesConsumed
			totalConsumed += response.bytesConsumed
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &common.RespValue{
				Type:  enums.ArrayRespType,
				Array: values,
			},
			bytesConsumed: totalConsumed,
		}

	default:
		return getParseErrorResp(common.ProtocolError("invalid RESP type"))
	}
}

func readLine(buffer []byte) int {
	for i := 0; i+1 < len(buffer); i++ {
		if buffer[i] == '\r' && buffer[i+1] == '\n' {
			return i
		}
	}
	return -1
}
