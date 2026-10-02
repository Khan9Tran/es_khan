package grpcclient

import (
	"fmt"
	"sort"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

// ParseProtoContent parses a raw .proto content string into ServiceInfo structures.
func ParseProtoContent(content string) ([]ServiceInfo, error) {
	if content == "" {
		return nil, fmt.Errorf("proto content cannot be empty")
	}

	files := map[string]string{
		"input.proto": content,
	}

	return ParseProtoFiles(files, "input.proto")
}

// ParseProtoFiles parses a map of filename -> content into ServiceInfo structures.
func ParseProtoFiles(files map[string]string, rootFiles ...string) ([]ServiceInfo, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no proto files provided")
	}

	parser := protoparse.Parser{
		Accessor: protoparse.FileContentsFromMap(files),
	}

	if len(rootFiles) == 0 {
		for f := range files {
			rootFiles = append(rootFiles, f)
		}
		sort.Strings(rootFiles)
	}

	fileDescs, err := parser.ParseFiles(rootFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto: %w", err)
	}

	var results []ServiceInfo
	for _, fd := range fileDescs {
		for _, svc := range fd.GetServices() {
			svcInfo := extractServiceInfo(svc)
			results = append(results, svcInfo)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func extractServiceInfo(svc *desc.ServiceDescriptor) ServiceInfo {
	var methods []MethodInfo
	for _, m := range svc.GetMethods() {
		methodInfo := MethodInfo{
			Name:              m.GetName(),
			FullName:          m.GetFullyQualifiedName(),
			IsClientStreaming: m.IsClientStreaming(),
			IsServerStreaming: m.IsServerStreaming(),
			InputType:         m.GetInputType().GetFullyQualifiedName(),
			OutputType:        m.GetOutputType().GetFullyQualifiedName(),
			MockRequest:       GenerateMockJSON(m.GetInputType()),
		}
		methods = append(methods, methodInfo)
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return ServiceInfo{
		Name:    svc.GetFullyQualifiedName(),
		Methods: methods,
	}
}
