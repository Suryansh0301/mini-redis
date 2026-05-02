package common

import (
	"fmt"

	"github.com/suryansh0301/mini-redis/internal/enums"
)

type RespValue struct {
	Type   enums.RespType
	Str    string
	Int    int64
	Array  []*RespValue
	IsNull bool
}

func (r *RespValue) IsType(t enums.RespType) bool {
	return r.Type == t
}

func (r *RespValue) IsEmpty() bool {
	switch r.Type {
	case enums.ArrayRespType:
		return len(r.Array) == 0
	case enums.SimpleStringRespType:
		return len(r.Str) == 0
	case enums.BulkStringRespType:
		return len(r.Str) == 0
	case enums.ErrorRespType:
		return len(r.Str) == 0
	default:
		return false
	}
}

func WrongNumberOfArgumentsError(command string) string {
	return fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)
}
