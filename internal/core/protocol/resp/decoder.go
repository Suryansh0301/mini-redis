package resp

import (
	"strings"

	"github.com/suryansh0301/mini-redis/internal/core/commands"
	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func Decoder(parsedResp ParseResp) (commands.Command, error) {
	if parsedResp.resp == nil {
		return commands.Command{}, common.ProtocolError("empty request")
	}

	if parsedResp.resp.Type != enums.ArrayRespType {
		return commands.Command{}, common.ProtocolError("command must be array")
	}

	if parsedResp.resp.IsNull {
		return commands.Command{}, common.ProtocolError("null array not allowed")
	}

	if len(parsedResp.resp.Array) == 0 {
		return commands.Command{}, common.ProtocolError("empty command array")
	}

	// Command name
	cmdValue := parsedResp.resp.Array[0]

	if cmdValue.Type != enums.BulkStringRespType {
		return commands.Command{}, common.ProtocolError("command name must be bulk string")
	}

	if cmdValue.IsNull || len(cmdValue.Str) == 0 {
		return commands.Command{}, common.ProtocolError("invalid command name")
	}

	commandName := strings.ToUpper(cmdValue.Str)

	// Arguments
	args := make([]string, 0, len(parsedResp.resp.Array)-1)

	for i := 1; i < len(parsedResp.resp.Array); i++ {
		arg := parsedResp.resp.Array[i]

		if arg.Type != enums.BulkStringRespType {
			return commands.Command{}, common.ProtocolError("arguments must be bulk strings")
		}

		if arg.IsNull {
			return commands.Command{}, common.ProtocolError("null argument not allowed")
		}

		args = append(args, arg.Str)
	}

	return commands.Command{
		Name: commandName,
		Args: args,
	}, nil
}
