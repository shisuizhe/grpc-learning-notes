package serializer

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func ProtobufToJSON(message proto.Message) (string, error) {
	marshaler := jsonpb.Marshaler{
		EnumsAsInts: false, // 是否将枚举值呈现为整数
		EmitDefaults: true, // 是否使用零值呈现字段
		Indent: "  ", 		// 使用什么缩进
		OrigName: true,		// 是否对字段使用原始文件中定义的原始字段名称
	}
	return marshaler.MarshalToString(message)
}
