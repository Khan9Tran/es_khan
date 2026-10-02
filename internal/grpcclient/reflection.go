package grpcclient

import (
	"context"
	"fmt"
	"sort"

	"github.com/jhump/protoreflect/grpcreflect"
)

// Internal reflection service names to omit from user view
var internalServices = map[string]bool{
	"grpc.reflection.v1alpha.ServerReflection": true,
	"grpc.reflection.v1.ServerReflection":      true,
}

// DiscoverServices connects to a gRPC target and discovers all services and methods via Server Reflection.
func DiscoverServices(ctx context.Context, cfg TargetConfig) ([]ServiceInfo, error) {
	conn, err := DialTarget(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := grpcreflect.NewClientAuto(ctx, conn)
	defer client.Reset()

	serviceNames, err := client.ListServices()
	if err != nil {
		return nil, fmt.Errorf("reflection error from %s: %w", cfg.Target, err)
	}

	var results []ServiceInfo
	for _, svcName := range serviceNames {
		if internalServices[svcName] {
			continue
		}

		svcDesc, err := client.ResolveService(svcName)
		if err != nil {
			// If a service fails to resolve, still report it with empty methods
			results = append(results, ServiceInfo{
				Name:    svcName,
				Methods: []MethodInfo{},
			})
			continue
		}

		var methods []MethodInfo
		for _, m := range svcDesc.GetMethods() {
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

		// Sort methods by name
		sort.Slice(methods, func(i, j int) bool {
			return methods[i].Name < methods[j].Name
		})

		results = append(results, ServiceInfo{
			Name:    svcName,
			Methods: methods,
		})
	}

	// Sort services by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}
