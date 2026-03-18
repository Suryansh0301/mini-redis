package resp

import (
	"fmt"
	"strconv"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

var encoderHandler map[enums.RespType]func(value common.RespValue) []byte

func init() {
	encoderHandler = make(map[enums.RespType]func(value common.RespValue) []byte)

	encoderHandler[enums.SimpleStringRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf("+" + value.Str + "\r\n")
		return []byte(response)
	}

	encoderHandler[enums.IntRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf(":" + strconv.FormatInt(value.Int, 10) + "\r\n")
		return []byte(response)
	}

	encoderHandler[enums.BulkStringRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf("$" + (value.Str) + "\r\n" + value.Str + "\r\n")
		return []byte(response)
	}

	encoderHandler[enums.ErrorRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf("-" + value.Str + "\r\n")
		return []byte(response)
	}
}

func Encoder(resp common.RespValue) []byte {
	return encoderHandler[resp.Type](resp)
}
