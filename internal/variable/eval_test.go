package variable

import (
	"strconv"
	"strings"
	"testing"
)

func TestEval(t *testing.T) {
	vars := map[string]string{
		"base_url": "https://api.example.com/v1",
		"user_id":  "42",
		"token":    "secret-jwt-token",
	}

	tests := []struct {
		name     string
		input    string
		contains string
		check    func(string) bool
	}{
		{
			name:     "User variables",
			input:    "{{base_url}}/users/{{user_id}}",
			contains: "https://api.example.com/v1/users/42",
		},
		{
			name:  "Builtin timestamp",
			input: "Time: {{$timestamp}}",
			check: func(s string) bool {
				parts := strings.Split(s, "Time: ")
				if len(parts) != 2 {
					return false
				}
				ts, err := strconv.ParseInt(parts[1], 10, 64)
				return err == nil && ts > 1000000000
			},
		},
		{
			name:  "Builtin UUID",
			input: "UUID: {{$uuid}}",
			check: func(s string) bool {
				uuid := strings.TrimPrefix(s, "UUID: ")
				return len(uuid) == 36 && strings.Count(uuid, "-") == 4
			},
		},
		{
			name:  "Builtin date",
			input: "Date: {{$date}}",
			check: func(s string) bool {
				d := strings.TrimPrefix(s, "Date: ")
				return len(d) == 10 && d[4] == '-' && d[7] == '-'
			},
		},
		{
			name:  "Builtin randomInt",
			input: "Rand: {{$randomInt}}",
			check: func(s string) bool {
				r := strings.TrimPrefix(s, "Rand: ")
				num, err := strconv.Atoi(r)
				return err == nil && num >= 100000 && num <= 999999
			},
		},
		{
			name:     "Combined user & builtin",
			input:    "Bearer {{token}} at {{$date}}",
			contains: "Bearer secret-jwt-token at ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Eval(tt.input, vars)
			if tt.contains != "" && !strings.Contains(res, tt.contains) {
				t.Errorf("Eval(%q) = %q; expected to contain %q", tt.input, res, tt.contains)
			}
			if tt.check != nil && !tt.check(res) {
				t.Errorf("Eval(%q) = %q; check failed", tt.input, res)
			}
		})
	}
}

func TestEvalMapAndSlice(t *testing.T) {
	vars := map[string]string{
		"env": "prod",
		"ver": "v2",
	}

	headers := map[string]string{
		"X-Env":     "{{env}}",
		"X-Version": "{{ver}}",
	}
	evaluatedHeaders := EvalMap(headers, vars)
	if evaluatedHeaders["X-Env"] != "prod" || evaluatedHeaders["X-Version"] != "v2" {
		t.Errorf("EvalMap failed, got %v", evaluatedHeaders)
	}

	slice := []string{"api/{{env}}", "{{ver}}/docs"}
	evaluatedSlice := EvalSlice(slice, vars)
	if evaluatedSlice[0] != "api/prod" || evaluatedSlice[1] != "v2/docs" {
		t.Errorf("EvalSlice failed, got %v", evaluatedSlice)
	}
}
