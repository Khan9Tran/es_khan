package grpcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Invoke executes a dynamic RPC call based on the provided InvokeRequest.
func Invoke(ctx context.Context, req InvokeRequest) (*InvokeResponse, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target is required")
	}
	if req.Service == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if req.Method == "" {
		return nil, fmt.Errorf("method name is required")
	}

	conn, err := DialTarget(ctx, TargetConfig{
		Target:             req.Target,
		Plaintext:          req.Plaintext,
		InsecureSkipVerify: req.InsecureSkipVerify,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Resolve the service and method descriptor via reflection
	reflectClient := grpcreflect.NewClientAuto(ctx, conn)
	defer reflectClient.Reset()

	svcDesc, err := reflectClient.ResolveService(req.Service)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service %q via reflection: %w", req.Service, err)
	}

	methodDesc := svcDesc.FindMethodByName(req.Method)
	if methodDesc == nil {
		return nil, fmt.Errorf("method %q not found in service %q", req.Method, req.Service)
	}

	// Prepare dynamic request message from input JSON
	inputDesc := methodDesc.GetInputType()
	dynMsg := dynamic.NewMessage(inputDesc)

	bodyBytes := []byte(strings.TrimSpace(req.Body))
	if len(bodyBytes) == 0 {
		bodyBytes = []byte("{}")
	}

	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err := dynMsg.UnmarshalJSONPB(unmarshaler, bodyBytes); err != nil {
		return nil, fmt.Errorf("invalid JSON payload for %s: %w", inputDesc.GetName(), err)
	}

	// Prepare metadata headers
	callCtx := ctx
	if len(req.Metadata) > 0 {
		md := metadata.New(req.Metadata)
		callCtx = metadata.NewOutgoingContext(callCtx, md)
	}

	// Set timeout
	timeout := 30 * time.Second
	if req.TimeoutMs > 0 {
		timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}
	var cancel context.CancelFunc
	callCtx, cancel = context.WithTimeout(callCtx, timeout)
	defer cancel()

	stub := grpcdynamic.NewStub(conn)
	var respHeaders metadata.MD
	var respTrailers metadata.MD

	start := time.Now()
	respMsg, rpcErr := stub.InvokeRpc(callCtx, methodDesc, dynMsg, grpc.Header(&respHeaders), grpc.Trailer(&respTrailers))
	tookMs := time.Since(start).Milliseconds()

	invResp := &InvokeResponse{
		Headers:  respHeaders,
		Trailers: respTrailers,
		TookMs:   tookMs,
	}

	if rpcErr != nil {
		st, ok := status.FromError(rpcErr)
		if ok {
			invResp.StatusCode = st.Code().String()
			invResp.Code = uint32(st.Code())
			invResp.Message = st.Message()
		} else {
			invResp.StatusCode = "UNKNOWN"
			invResp.Code = 2
			invResp.Message = rpcErr.Error()
		}
		return invResp, nil
	}

	invResp.StatusCode = "OK"
	invResp.Code = 0

	if dynResp, ok := respMsg.(*dynamic.Message); ok {
		marshaler := &jsonpb.Marshaler{
			Indent:       "  ",
			EmitDefaults: true,
		}
		rawJSON, err := dynResp.MarshalJSONPB(marshaler)
		if err == nil {
			invResp.RawJSON = string(rawJSON)
			var parsed any
			if err := json.Unmarshal(rawJSON, &parsed); err == nil {
				invResp.Response = parsed
			}
		}
	}

	return invResp, nil
}
