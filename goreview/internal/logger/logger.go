package logger

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Level represents logging levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured logger with secret masking
type Logger struct {
	level     Level
	output    io.Writer
	prefix    string
	fields    map[string]interface{}
	mu        sync.Mutex
	maskFuncs []MaskFunc
}

// MaskFunc is a function that masks sensitive data
type MaskFunc func(string) string

// Default secret patterns to mask
var defaultSecretPatterns = []*regexp.Regexp{
	// API Keys
	regexp.MustCompile(`(?i)(sk-[a-zA-Z0-9]{20,})`),                                                    // OpenAI
	regexp.MustCompile(`(?i)(AIza[a-zA-Z0-9_-]{35})`),                                                  // Google API
	regexp.MustCompile(`(?i)(gsk_[a-zA-Z0-9]{20,})`),                                                   // Groq
	regexp.MustCompile(`(?i)(ghp_[a-zA-Z0-9]{36})`),                                                    // GitHub PAT
	regexp.MustCompile(`(?i)(gho_[a-zA-Z0-9]{36})`),                                                    // GitHub OAuth
	regexp.MustCompile(`(?i)(ghs_[a-zA-Z0-9]{36})`),                                                    // GitHub App
	regexp.MustCompile(`(?i)(ghr_[a-zA-Z0-9]{36})`),                                                    // GitHub Refresh
	regexp.MustCompile(`(?i)(github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59})`),                             // GitHub Fine-grained
	regexp.MustCompile(`(?i)(xoxb-[a-zA-Z0-9-]+)`),                                                     // Slack Bot
	regexp.MustCompile(`(?i)(xoxp-[a-zA-Z0-9-]+)`),                                                     // Slack User
	regexp.MustCompile(`(?i)(AKIA[A-Z0-9]{16})`),                                                       // AWS Access Key
	regexp.MustCompile(`(?i)([a-zA-Z0-9/+]{40})`),                                                      // AWS Secret (40 chars base64-like)
	regexp.MustCompile(`(?i)(Bearer\s+[a-zA-Z0-9._-]+)`),                                               // Bearer tokens
	regexp.MustCompile(`(?i)(api[_-]?key[=:]\s*["']?[a-zA-Z0-9_-]{16,}["']?)`),                         // Generic API key
	regexp.MustCompile(`(?i)(secret[=:]\s*["']?[a-zA-Z0-9_-]{16,}["']?)`),                              // Generic secret
	regexp.MustCompile(`(?i)(password[=:]\s*["']?[^\s"']{8,}["']?)`),                                   // Passwords
	regexp.MustCompile(`(?i)(token[=:]\s*["']?[a-zA-Z0-9._-]{20,}["']?)`),                              // Generic tokens
	regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+ PRIVATE KEY-----`), // Private keys
}

// Sensitive field names that should be masked in structured logging
var sensitiveFieldNames = map[string]bool{
	"password":      true,
	"secret":        true,
	"token":         true,
	"api_key":       true,
	"apikey":        true,
	"api-key":       true,
	"private_key":   true,
	"privatekey":    true,
	"access_token":  true,
	"accesstoken":   true,
	"auth":          true,
	"authorization": true,
	"credential":    true,
	"credentials":   true,
}

var defaultLogger *Logger
var once sync.Once

// Default returns the default logger
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(LevelInfo, os.Stdout)
	})
	return defaultLogger
}

// New creates a new logger
func New(level Level, output io.Writer) *Logger {
	l := &Logger{
		level:  level,
		output: output,
		fields: make(map[string]interface{}),
	}
	l.maskFuncs = append(l.maskFuncs, l.maskPatterns)
	return l
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// WithField returns a new logger with the field added
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// WithFields returns a new logger with the fields added
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &Logger{
		level:     l.level,
		output:    l.output,
		prefix:    l.prefix,
		fields:    newFields,
		maskFuncs: l.maskFuncs,
	}
}

// WithPrefix returns a new logger with the prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:     l.level,
		output:    l.output,
		prefix:    prefix,
		fields:    l.fields,
		maskFuncs: l.maskFuncs,
	}
}

// AddMaskFunc adds a custom masking function
func (l *Logger) AddMaskFunc(fn MaskFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maskFuncs = append(l.maskFuncs, fn)
}

// maskPatterns masks known secret patterns
func (l *Logger) maskPatterns(s string) string {
	result := s
	for _, pattern := range defaultSecretPatterns {
		result = pattern.ReplaceAllStringFunc(result, maskString)
	}
	return result
}

// maskString masks a string showing only first and last 4 chars
func maskString(s string) string {
	if len(s) <= 8 {
		return "***MASKED***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

// mask applies all mask functions to a string
func (l *Logger) mask(s string) string {
	for _, fn := range l.maskFuncs {
		s = fn(s)
	}
	return s
}

// maskValue masks a value if it's a string and the key is sensitive
func (l *Logger) maskValue(key string, value interface{}) interface{} {
	keyLower := strings.ToLower(key)
	if sensitiveFieldNames[keyLower] {
		if str, ok := value.(string); ok {
			return maskString(str)
		}
		return "***MASKED***"
	}
	if str, ok := value.(string); ok {
		return l.mask(str)
	}
	return value
}

// formatFields formats the fields for output
func (l *Logger) formatFields() string {
	if len(l.fields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(l.fields))
	for k, v := range l.fields {
		maskedValue := l.maskValue(k, v)
		parts = append(parts, fmt.Sprintf("%s=%v", k, maskedValue))
	}
	return " " + strings.Join(parts, " ")
}

// log logs a message at the given level
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Mask the message
	formattedMsg = l.mask(formattedMsg)

	prefix := ""
	if l.prefix != "" {
		prefix = "[" + l.prefix + "] "
	}

	fields := l.formatFields()

	line := fmt.Sprintf("%s %s %s%s%s\n", timestamp, level.String(), prefix, formattedMsg, fields)
	fmt.Fprint(l.output, line)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// Debugf is an alias for Debug
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.Debug(msg, args...)
}

// Infof is an alias for Info
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.Info(msg, args...)
}

// Warnf is an alias for Warn
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.Warn(msg, args...)
}

// Errorf is an alias for Error
func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.Error(msg, args...)
}

// Package-level functions using default logger

// Debug logs a debug message using the default logger
func Debug(msg string, args ...interface{}) {
	Default().Debug(msg, args...)
}

// Info logs an info message using the default logger
func Info(msg string, args ...interface{}) {
	Default().Info(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...interface{}) {
	Default().Warn(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...interface{}) {
	Default().Error(msg, args...)
}

// SetLevel sets the level of the default logger
func SetLevel(level Level) {
	Default().SetLevel(level)
}

// SetOutput sets the output of the default logger
func SetOutput(w io.Writer) {
	Default().SetOutput(w)
}

// WithField returns a new logger with the field added
func WithField(key string, value interface{}) *Logger {
	return Default().WithField(key, value)
}

// WithFields returns a new logger with the fields added
func WithFields(fields map[string]interface{}) *Logger {
	return Default().WithFields(fields)
}

// MaskSecrets masks all known secret patterns in a string
func MaskSecrets(s string) string {
	return Default().mask(s)
}

// IsSensitiveKey checks if a key name is sensitive
func IsSensitiveKey(key string) bool {
	return sensitiveFieldNames[strings.ToLower(key)]
}
