package grpcclient

import (
	"encoding/json"

	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GenerateMockJSON creates an indented JSON string representing a sample payload for the given MessageDescriptor.
func GenerateMockJSON(md *desc.MessageDescriptor) string {
	if md == nil {
		return "{}"
	}

	visited := make(map[string]int)
	obj := generateMockMessage(md, visited, 0)
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func generateMockMessage(md *desc.MessageDescriptor, visited map[string]int, depth int) any {
	if depth > 4 || visited[md.GetFullyQualifiedName()] > 1 {
		return map[string]any{}
	}

	visited[md.GetFullyQualifiedName()]++
	defer func() {
		visited[md.GetFullyQualifiedName()]--
	}()

	result := make(map[string]any)
	for _, f := range md.GetFields() {
		fieldName := f.GetJSONName()
		if fieldName == "" {
			fieldName = f.GetName()
		}

		if f.IsMap() {
			valType := f.GetMapValueType()
			val := generateMockValue(valType, visited, depth+1)
			result[fieldName] = map[string]any{"key": val}
		} else if f.IsRepeated() {
			val := generateMockValue(f, visited, depth+1)
			result[fieldName] = []any{val}
		} else {
			result[fieldName] = generateMockValue(f, visited, depth+1)
		}
	}

	return result
}

func generateMockValue(f *desc.FieldDescriptor, visited map[string]int, depth int) any {
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return 0.0

	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return 0

	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return false

	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return ""

	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return ""

	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enumType := f.GetEnumType()
		if enumType != nil && len(enumType.GetValues()) > 0 {
			return enumType.GetValues()[0].GetName()
		}
		return 0

	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		msgType := f.GetMessageType()
		if msgType == nil {
			return map[string]any{}
		}
		// Special handling for well-known types if needed
		if msgType.GetFullyQualifiedName() == "google.protobuf.Timestamp" {
			return "2026-10-02T00:00:00Z"
		}
		return generateMockMessage(msgType, visited, depth)

	default:
		return ""
	}
}
