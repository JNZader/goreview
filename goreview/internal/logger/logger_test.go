package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestMaskSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OpenAI API key",
			input:    "Using key sk-1234567890abcdefghijklmnop",
			expected: "Using key sk-1***mnop",
		},
		{
			name:     "GitHub PAT",
			input:    "Token: ghp_1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "Token: ghp_***wxyz",
		},
		{
			name:     "Groq API key",
			input:    "GROQ_API_KEY=gsk_abcdefghijklmnopqrstuvwx",
			expected: "GROQ_API_KEY=gsk_***uvwx",
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bear***XVCJ9",
		},
		{
			name:     "Generic API key pattern",
			input:    "api_key=abcd1234567890efghij",
			expected: "api_***ghij",
		},
		{
			name:     "Private key block",
			input:    "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
			expected: "----***-----",
		},
		{
			name:     "No secrets",
			input:    "This is a normal log message",
			expected: "This is a normal log message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSecrets(tt.input)
			if !strings.Contains(result, "***") && strings.Contains(tt.expected, "***") {
				t.Errorf("Expected masked output, got: %s", result)
			}
			// Verify original secret is not present
			if strings.Contains(tt.expected, "***") && result == tt.input {
				t.Errorf("Secret was not masked: %s", result)
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	// Debug should not appear when level is Info
	log.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message should not appear when level is Info")
	}

	// Info should appear
	log.Info("info message")
	if !strings.Contains(buf.String(), "INFO") {
		t.Error("Info message should appear")
	}

	buf.Reset()

	// Warn should appear
	log.Warn("warn message")
	if !strings.Contains(buf.String(), "WARN") {
		t.Error("Warn message should appear")
	}

	buf.Reset()

	// Error should appear
	log.Error("error message")
	if !strings.Contains(buf.String(), "ERROR") {
		t.Error("Error message should appear")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	log.WithField("request_id", "abc123").Info("test message")

	output := buf.String()
	if !strings.Contains(output, "request_id=abc123") {
		t.Errorf("Expected field in output, got: %s", output)
	}
}

func TestLoggerMasksSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	log.WithField("password", "super_secret_password").Info("user login")

	output := buf.String()
	if strings.Contains(output, "super_secret_password") {
		t.Error("Password should be masked")
	}
	if !strings.Contains(output, "***MASKED***") && !strings.Contains(output, "supe***word") {
		t.Errorf("Expected masked password in output, got: %s", output)
	}
}

func TestLoggerMasksSecretsInMessage(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	log.Info("Using API key sk-1234567890abcdefghijklmnop for request")

	output := buf.String()
	if strings.Contains(output, "sk-1234567890abcdefghijklmnop") {
		t.Error("API key should be masked in message")
	}
}

func TestLoggerWithPrefix(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	log.WithPrefix("ENGINE").Info("starting review")

	output := buf.String()
	if !strings.Contains(output, "[ENGINE]") {
		t.Errorf("Expected prefix in output, got: %s", output)
	}
}

func TestIsSensitiveKey(t *testing.T) {
	sensitiveKeys := []string{
		"password", "PASSWORD", "Password",
		"secret", "api_key", "apikey", "token",
		"private_key", "access_token", "authorization",
	}

	for _, key := range sensitiveKeys {
		if !IsSensitiveKey(key) {
			t.Errorf("Expected %s to be sensitive", key)
		}
	}

	nonSensitiveKeys := []string{
		"username", "email", "id", "name", "status",
	}

	for _, key := range nonSensitiveKeys {
		if IsSensitiveKey(key) {
			t.Errorf("Expected %s to NOT be sensitive", key)
		}
	}
}

func TestLoggerFormatting(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	log.Info("Processing file %s with %d lines", "main.go", 100)

	output := buf.String()
	if !strings.Contains(output, "Processing file main.go with 100 lines") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func BenchmarkLoggerMasking(b *testing.B) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	message := "Using API key sk-1234567890abcdefghijklmnop and token ghp_1234567890abcdefghijklmnopqrstuvwxyz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		log.Info("%s", message)
	}
}

func BenchmarkLoggerNoMasking(b *testing.B) {
	var buf bytes.Buffer
	log := New(LevelInfo, &buf)

	message := "Processing file main.go with 100 lines in directory src"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		log.Info("%s", message)
	}
}
