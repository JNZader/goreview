package providers

import "strings"

// ReviewMode represents a specialized review focus mode.
type ReviewMode string

const (
	// ModeDefault performs a balanced review across all aspects.
	ModeDefault ReviewMode = "default"

	// ModeSecurity focuses on security vulnerabilities and OWASP Top 10.
	ModeSecurity ReviewMode = "security"

	// ModePerformance focuses on performance issues, N+1 queries, memory leaks.
	ModePerformance ReviewMode = "perf"

	// ModeClean focuses on clean code principles, SOLID, DRY, naming.
	ModeClean ReviewMode = "clean"

	// ModeDocs focuses on missing documentation, comments, JSDoc/GoDoc.
	ModeDocs ReviewMode = "docs"

	// ModeTests focuses on test coverage, edge cases, mocking issues.
	ModeTests ReviewMode = "tests"
)

// ModePrompts contains the mode-specific instructions for the reviewer.
// These prompts define WHAT to analyze (focus area), not HOW to communicate (that's personality).
var ModePrompts = map[ReviewMode]string{
	ModeDefault: `Focus on:
- Real bugs and logic errors
- Security vulnerabilities
- Performance issues
- Code clarity and maintainability

Report only issues that matter, not nitpicks.`,

	ModeSecurity: `SECURITY REVIEW MODE - Focus exclusively on security vulnerabilities:

CHECK FOR:
- Injection vulnerabilities (SQL, NoSQL, command, LDAP, XPath, XSS)
- Broken authentication and session management
- Sensitive data exposure (hardcoded secrets, API keys, passwords)
- XML External Entities (XXE)
- Broken access control (IDOR, privilege escalation)
- Security misconfiguration
- Insecure deserialization
- Using components with known vulnerabilities
- Insufficient logging and monitoring
- SSRF (Server-Side Request Forgery)

ALSO CHECK:
- Input validation and sanitization
- Output encoding
- Cryptographic weaknesses (weak algorithms, improper key management)
- Race conditions leading to security issues
- Path traversal vulnerabilities
- Insecure file operations
- Missing security headers
- CORS misconfigurations

SEVERITY GUIDELINES:
- CRITICAL: RCE, SQL injection, auth bypass, secret exposure
- ERROR: XSS, IDOR, SSRF, path traversal
- WARNING: Missing validation, weak crypto, verbose errors
- INFO: Missing security headers, logging gaps

Only report security-related issues. Ignore style, performance, or documentation issues.`,

	ModePerformance: `PERFORMANCE REVIEW MODE - Focus exclusively on performance issues:

CHECK FOR:
- N+1 query problems
- Missing database indexes (queries on unindexed fields)
- Unnecessary database queries in loops
- Memory leaks and unbounded allocations
- Inefficient algorithms (O(n²) when O(n) is possible)
- Blocking operations in async code
- Missing caching opportunities
- Large object allocations in hot paths
- Unnecessary string concatenations in loops
- Redundant computations
- Missing pagination for large datasets
- Synchronous I/O in async contexts

ALSO CHECK:
- Connection pool exhaustion risks
- Missing timeouts on external calls
- Unbounded goroutines/threads
- Large payload serialization
- Inefficient regex patterns
- Missing lazy loading
- Unnecessary deep copies
- Hot path optimizations

SEVERITY GUIDELINES:
- CRITICAL: Memory leaks, connection leaks, O(n²) on large datasets
- ERROR: N+1 queries, blocking in async, missing timeouts
- WARNING: Suboptimal algorithms, missing caching
- INFO: Minor optimizations, documentation of performance notes

Only report performance-related issues. Ignore security, style, or documentation issues.`,

	ModeClean: `CLEAN CODE REVIEW MODE - Focus on code quality and maintainability:

CHECK FOR:
- SOLID principle violations:
  - Single Responsibility: classes/functions doing too much
  - Open/Closed: code that requires modification instead of extension
  - Liskov Substitution: improper inheritance hierarchies
  - Interface Segregation: large interfaces forcing empty implementations
  - Dependency Inversion: high-level modules depending on low-level details
- DRY violations (duplicated code that should be abstracted)
- YAGNI violations (unnecessary complexity for hypothetical futures)
- Poor naming (unclear variable/function/class names)
- Code smells:
  - Long methods (>20 lines typically)
  - Long parameter lists (>4 parameters)
  - Feature envy (method uses other class's data excessively)
  - Data clumps (same group of data appearing together)
  - Primitive obsession (using primitives instead of small objects)
  - Switch statements that should be polymorphism
  - Parallel inheritance hierarchies
  - Lazy class (class doing too little)
  - Speculative generality
  - Dead code

ALSO CHECK:
- Magic numbers/strings
- Deep nesting (>3 levels)
- Complex conditionals that need extraction
- Missing early returns
- Inconsistent error handling patterns
- Tight coupling between modules
- God classes/functions

SEVERITY GUIDELINES:
- ERROR: Major SOLID violations, high code duplication
- WARNING: Code smells, poor naming, deep nesting
- INFO: Style inconsistencies, minor improvements

Only report clean code issues. Ignore security or performance unless they're also maintainability problems.`,

	ModeDocs: `DOCUMENTATION REVIEW MODE - Focus on missing or incomplete documentation:

CHECK FOR:
- Missing function/method documentation:
  - Public APIs without descriptions
  - Complex functions without explanations
  - Non-obvious parameter purposes
  - Missing return value documentation
  - Missing error documentation
- Missing type/class documentation:
  - Undocumented public types
  - Missing field descriptions
  - Unclear struct/class purposes
- Missing package/module documentation
- Outdated or incorrect documentation
- Missing examples for complex usage
- Undocumented edge cases or limitations

FORMAT EXPECTATIONS BY LANGUAGE:
- Go: GoDoc comments (// FunctionName does...)
- JavaScript/TypeScript: JSDoc comments (/** @param @returns */)
- Python: Docstrings ("""Description""")
- Java: Javadoc (/** @param @return @throws */)
- Rust: Doc comments (/// or //!)

ALSO CHECK:
- Missing TODO/FIXME explanations
- Commented code without explanation
- Complex algorithms without explanation
- Missing API documentation
- Missing configuration documentation
- Unclear error messages that need context

SEVERITY GUIDELINES:
- ERROR: Missing docs on public APIs, incorrect documentation
- WARNING: Missing docs on complex internal functions
- INFO: Missing docs on simple private functions

Only report documentation issues. Ignore security, performance, or code structure.`,

	ModeTests: `TEST COVERAGE REVIEW MODE - Focus on testing gaps and test quality:

CHECK FOR:
- Missing test coverage:
  - Untested public functions/methods
  - Untested error paths
  - Untested edge cases (null, empty, boundary values)
  - Untested async/concurrent scenarios
  - Missing integration tests for complex flows
- Test quality issues:
  - Tests without assertions
  - Tests with too many assertions (test one thing)
  - Flaky tests (time-dependent, order-dependent)
  - Tests that test implementation, not behavior
  - Missing test descriptions/names
  - Duplicated test setup
  - Hard-to-understand test logic
- Mocking issues:
  - Over-mocking (mocking everything)
  - Under-mocking (tests hitting real services)
  - Incorrect mock setup
  - Mocks not verified

ALSO CHECK:
- Missing negative test cases
- Missing boundary condition tests
- Missing concurrency tests
- Missing error injection tests
- Test data management issues
- Missing test cleanup
- Assertions on wrong values
- Ignoring returned errors in tests

SEVERITY GUIDELINES:
- ERROR: Untested critical paths, tests without assertions
- WARNING: Missing edge case tests, flaky test patterns
- INFO: Test organization improvements, naming suggestions

Only report testing-related issues. Ignore production code style or documentation.`,
}

// ValidModes returns all valid mode names.
func ValidModes() []string {
	return []string{
		string(ModeDefault),
		string(ModeSecurity),
		string(ModePerformance),
		string(ModeClean),
		string(ModeDocs),
		string(ModeTests),
	}
}

// IsValidMode checks if a mode name is valid.
func IsValidMode(name string) bool {
	for _, m := range ValidModes() {
		if m == name {
			return true
		}
	}
	return false
}

// GetModePrompt returns the prompt for a given mode.
func GetModePrompt(name string) string {
	m := ReviewMode(name)
	if prompt, ok := ModePrompts[m]; ok {
		return prompt
	}
	return ModePrompts[ModeDefault]
}

// ParseModes parses a comma-separated list of modes and returns valid ones.
// Example: "security,perf" -> ["security", "perf"]
func ParseModes(modesStr string) []ReviewMode {
	if modesStr == "" || modesStr == "default" {
		return []ReviewMode{ModeDefault}
	}

	parts := strings.Split(modesStr, ",")
	modes := make([]ReviewMode, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if IsValidMode(p) && p != "default" {
			modes = append(modes, ReviewMode(p))
		}
	}

	if len(modes) == 0 {
		return []ReviewMode{ModeDefault}
	}

	return modes
}

// CombineModePrompts combines multiple mode prompts into one.
func CombineModePrompts(modes []ReviewMode) string {
	if len(modes) == 0 || (len(modes) == 1 && modes[0] == ModeDefault) {
		return ModePrompts[ModeDefault]
	}

	var prompts []string
	for _, m := range modes {
		if prompt, ok := ModePrompts[m]; ok {
			prompts = append(prompts, prompt)
		}
	}

	if len(prompts) == 0 {
		return ModePrompts[ModeDefault]
	}

	return "MULTI-MODE REVIEW: Apply all the following review focuses:\n\n" +
		strings.Join(prompts, "\n\n---\n\n")
}
