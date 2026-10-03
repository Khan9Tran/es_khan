package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"eskhan/internal/variable"
)

const (
	// DefaultTimeout is 30 seconds if not specified by user.
	DefaultTimeout = 30 * time.Second
	// MaxResponseBodySize is 50MB.
	MaxResponseBodySize = 50 * 1024 * 1024
)

// Execute executes an HTTP request with CORS bypass, detailed network tracing, and response parsing.
func Execute(ctx context.Context, req Request) (*Response, error) {
	// Variable substitution if variables are provided or dynamic variables are present
	req.URL = variable.Eval(req.URL, req.Variables)
	req.QueryParams = variable.EvalMap(req.QueryParams, req.Variables)
	req.Headers = variable.EvalMap(req.Headers, req.Variables)
	req.Body = variable.Eval(req.Body, req.Variables)
	req.FormData = variable.EvalMap(req.FormData, req.Variables)

	if req.URL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	// 1. Parse and build target URL with Query Params
	targetURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", req.URL, err)
	}

	// Default to http scheme if missing
	if targetURL.Scheme == "" {
		targetURL.Scheme = "http"
	}

	if len(req.QueryParams) > 0 {
		q := targetURL.Query()
		for k, v := range req.QueryParams {
			if strings.TrimSpace(k) != "" {
				q.Set(k, v)
			}
		}
		targetURL.RawQuery = q.Encode()
	}

	// 2. Prepare Request Body & Content-Type
	var bodyReader io.Reader
	contentType := ""

	switch strings.ToLower(req.BodyType) {
	case "json":
		bodyReader = strings.NewReader(req.Body)
		contentType = "application/json"

	case "x_www_form_urlencoded":
		formValues := url.Values{}
		if len(req.FormData) > 0 {
			for k, v := range req.FormData {
				formValues.Set(k, v)
			}
		} else if req.Body != "" {
			parsedVals, err := url.ParseQuery(req.Body)
			if err == nil {
				formValues = parsedVals
			}
		}
		encoded := formValues.Encode()
		bodyReader = strings.NewReader(encoded)
		contentType = "application/x-www-form-urlencoded"

	case "form_data":
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for k, v := range req.FormData {
			_ = w.WriteField(k, v)
		}
		_ = w.Close()
		bodyReader = &b
		contentType = w.FormDataContentType()

	case "raw":
		bodyReader = strings.NewReader(req.Body)

	default:
		if req.Body != "" {
			bodyReader = strings.NewReader(req.Body)
		}
	}

	// 3. Setup context with timeout
	timeout := DefaultTimeout
	if req.TimeoutMs > 0 {
		timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 4. Setup Network Tracing via net/http/httptrace
	var timings NetworkTimings
	var dnsStart, connectStart, tlsStart time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				timings.DNSLookupMs = time.Since(dnsStart).Milliseconds()
			}
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if !connectStart.IsZero() {
				timings.TCPConnectMs = time.Since(connectStart).Milliseconds()
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				timings.TLSHandshakeMs = time.Since(tlsStart).Milliseconds()
			}
		},
		GotFirstResponseByte: func() {
			// TTFB measured from start of request execution
		},
	}
	traceCtx := httptrace.WithClientTrace(execCtx, trace)

	httpReq, err := http.NewRequestWithContext(traceCtx, method, targetURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Apply headers
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	for k, v := range req.Headers {
		if strings.TrimSpace(k) != "" {
			httpReq.Header.Set(k, v)
		}
	}

	// 5. Setup Custom Transport & Client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: req.InsecureSkipVerify,
		},
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	if req.FollowRedirects != nil && !*req.FollowRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 6. Execute Request
	startTime := time.Now()
	resp, reqErr := httpClient.Do(httpReq)
	totalDuration := time.Since(startTime).Milliseconds()
	timings.TotalMs = totalDuration

	if reqErr != nil {
		// Return response with error details so user can inspect timing & failure
		return &Response{
			StatusCode: 0,
			StatusText: "Error",
			Headers:    make(map[string][]string),
			Timings:    timings,
			Error:      reqErr.Error(),
		}, nil
	}
	defer resp.Body.Close()

	// 7. Read Response Body
	limitReader := io.LimitReader(resp.Body, MaxResponseBodySize)
	respBytes, readErr := io.ReadAll(limitReader)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	bodyStr := string(respBytes)
	sizeBytes := int64(len(respBytes))

	// Approximate TTFB as total minus body read time or record total
	timings.TTFBMs = timings.TotalMs

	return &Response{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Headers:    resp.Header,
		Body:       bodyStr,
		Size:       sizeBytes,
		Timings:    timings,
	}, nil
}
