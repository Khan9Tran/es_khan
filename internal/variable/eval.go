package variable

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Eval replaces template variables like {{var}}, {{$timestamp}}, {{$uuid}}, {{$date}}, {{$randomInt}}.
func Eval(input string, userVars map[string]string) string {
	if input == "" {
		return ""
	}

	res := input
	now := time.Now()

	// 1. Built-in dynamic variables
	if strings.Contains(res, "{{$") {
		res = strings.ReplaceAll(res, "{{$timestamp}}", strconv.FormatInt(now.Unix(), 10))
		res = strings.ReplaceAll(res, "{{$timestamp_ms}}", strconv.FormatInt(now.UnixMilli(), 10))
		res = strings.ReplaceAll(res, "{{$date}}", now.Format("2006-01-02"))
		res = strings.ReplaceAll(res, "{{$datetime}}", now.Format(time.RFC3339))
		res = strings.ReplaceAll(res, "{{$isoTimestamp}}", now.UTC().Format("2006-01-02T15:04:05.000Z"))

		if strings.Contains(res, "{{$uuid}}") {
			// Generate pseudo-random UUID v4 format
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			b[6] = (b[6] & 0x0f) | 0x40 // version 4
			b[8] = (b[8] & 0x3f) | 0x80 // variant RFC4122
			uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
			res = strings.ReplaceAll(res, "{{$uuid}}", uuid)
		}

		if strings.Contains(res, "{{$randomInt}}") {
			n, err := rand.Int(rand.Reader, big.NewInt(900000))
			rnd := 100000
			if err == nil {
				rnd += int(n.Int64())
			}
			res = strings.ReplaceAll(res, "{{$randomInt}}", strconv.Itoa(rnd))
		}
	}

	// 2. User defined variables
	for k, v := range userVars {
		if k != "" {
			res = strings.ReplaceAll(res, "{{"+k+"}}", v)
		}
	}

	return res
}

// EvalMap replaces variables in keys and values of a map.
func EvalMap(m map[string]string, userVars map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		newKey := Eval(k, userVars)
		newVal := Eval(v, userVars)
		result[newKey] = newVal
	}
	return result
}

// EvalSlice replaces variables in each element of a string slice.
func EvalSlice(s []string, userVars map[string]string) []string {
	if s == nil {
		return nil
	}
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = Eval(v, userVars)
	}
	return result
}
