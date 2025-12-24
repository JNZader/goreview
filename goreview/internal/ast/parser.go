// Package ast provides Abstract Syntax Tree parsing for code analysis.
// It extracts structural context to improve LLM code review accuracy.
package ast

import (
	"regexp"
	"strings"
)

// Context represents the extracted context from code
type Context struct {
	// Package/Module information
	Package string `json:"package,omitempty"`
	Module  string `json:"module,omitempty"`

	// Imports
	Imports []Import `json:"imports,omitempty"`

	// Definitions in the file
	Functions  []Function  `json:"functions,omitempty"`
	Classes    []Class     `json:"classes,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
	Variables  []Variable  `json:"variables,omitempty"`
	Constants  []Variable  `json:"constants,omitempty"`

	// File metadata
	Language string `json:"language"`
	FilePath string `json:"file_path"`
}

// Import represents an import statement
type Import struct {
	Path  string `json:"path"`
	Alias string `json:"alias,omitempty"`
}

// Function represents a function definition
type Function struct {
	Name       string   `json:"name"`
	Receiver   string   `json:"receiver,omitempty"` // For methods
	Parameters []Param  `json:"parameters,omitempty"`
	Returns    []string `json:"returns,omitempty"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	IsExported bool     `json:"is_exported"`
	DocComment string   `json:"doc_comment,omitempty"`
}

// Param represents a function parameter
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Class represents a class/struct definition
type Class struct {
	Name       string   `json:"name"`
	Fields     []Field  `json:"fields,omitempty"`
	Methods    []string `json:"methods,omitempty"` // Method names
	Extends    string   `json:"extends,omitempty"`
	Implements []string `json:"implements,omitempty"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	IsExported bool     `json:"is_exported"`
}

// Field represents a class/struct field
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tags string `json:"tags,omitempty"` // For Go struct tags
}

// Interface represents an interface definition
type Interface struct {
	Name       string   `json:"name"`
	Methods    []string `json:"methods,omitempty"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	IsExported bool     `json:"is_exported"`
}

// Variable represents a variable/constant declaration
type Variable struct {
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	Line       int    `json:"line"`
	IsExported bool   `json:"is_exported"`
}

// Parser parses source code to extract AST context
type Parser struct {
	language string
}

// NewParser creates a new parser for the given language
func NewParser(language string) *Parser {
	return &Parser{
		language: strings.ToLower(language),
	}
}

// Parse extracts context from source code
func (p *Parser) Parse(code, filePath string) (*Context, error) {
	ctx := &Context{
		Language: p.language,
		FilePath: filePath,
	}

	lines := strings.Split(code, "\n")

	switch p.language {
	case "go", "golang":
		p.parseGo(lines, ctx)
	case "javascript", "js":
		p.parseJavaScript(lines, ctx)
	case "typescript", "ts":
		p.parseTypeScript(lines, ctx)
	case "python", "py":
		p.parsePython(lines, ctx)
	case "java":
		p.parseJava(lines, ctx)
	case "rust", "rs":
		p.parseRust(lines, ctx)
	default:
		// Generic parsing
		p.parseGeneric(lines, ctx)
	}

	return ctx, nil
}

// ParseDiff extracts context from a diff, focusing on changed areas
func (p *Parser) ParseDiff(diff, fullContent, filePath string) (*DiffContext, error) {
	// Parse the full file context
	fullCtx, err := p.Parse(fullContent, filePath)
	if err != nil {
		return nil, err
	}

	// Identify which functions/classes were modified
	changedLines := extractChangedLines(diff)

	dc := &DiffContext{
		FullContext:      fullCtx,
		ChangedFunctions: []Function{},
		ChangedClasses:   []Class{},
		SurroundingCode:  "",
	}

	// Find functions that contain changed lines
	for _, fn := range fullCtx.Functions {
		for _, line := range changedLines {
			if line >= fn.StartLine && line <= fn.EndLine {
				dc.ChangedFunctions = append(dc.ChangedFunctions, fn)
				break
			}
		}
	}

	// Find classes that contain changed lines
	for _, cls := range fullCtx.Classes {
		for _, line := range changedLines {
			if line >= cls.StartLine && line <= cls.EndLine {
				dc.ChangedClasses = append(dc.ChangedClasses, cls)
				break
			}
		}
	}

	return dc, nil
}

// DiffContext provides context specifically for a diff
type DiffContext struct {
	FullContext      *Context   `json:"full_context"`
	ChangedFunctions []Function `json:"changed_functions"`
	ChangedClasses   []Class    `json:"changed_classes"`
	SurroundingCode  string     `json:"surrounding_code,omitempty"`
}

// extractChangedLines extracts line numbers that were changed from a diff
func extractChangedLines(diff string) []int {
	var lines []int
	linePattern := regexp.MustCompile(`^@@\s*-\d+(?:,\d+)?\s*\+(\d+)(?:,(\d+))?\s*@@`)

	for _, line := range strings.Split(diff, "\n") {
		if matches := linePattern.FindStringSubmatch(line); len(matches) > 1 {
			// Parse the starting line number
			var startLine int
			parseIntSafe(matches[1], &startLine)
			lines = append(lines, startLine)
		}
	}

	return lines
}

func parseIntSafe(s string, result *int) {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	*result = n
}

// goParseState holds parsing state for Go files
type goParseState struct {
	ctx           *Context
	inImportBlock bool
	inTypeBlock   bool
	currentType   string
	typeStartLine int
	braceCount    int
}

// Go parsing patterns
var (
	goPackagePattern   = regexp.MustCompile(`^package\s+(\w+)`)
	goImportPattern    = regexp.MustCompile(`^\s*(?:import\s+)?(?:(\w+)\s+)?"([^"]+)"`)
	goImportBlockStart = regexp.MustCompile(`^import\s*\(`)
	goFuncPattern      = regexp.MustCompile(`^func\s+(?:\((\w+)\s+[^)]+\)\s+)?(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\)|(\w+))?\s*\{?`)
	goTypePattern      = regexp.MustCompile(`^type\s+(\w+)\s+(struct|interface)\s*\{?`)
	goVarPattern       = regexp.MustCompile(`^(?:var|const)\s+(\w+)\s+(\w+)?(?:\s*=\s*(.+))?`)
)

// Go-specific parsing
func (p *Parser) parseGo(lines []string, ctx *Context) {
	state := &goParseState{ctx: ctx}

	for i, line := range lines {
		lineNum := i + 1
		state.parseLine(line, lineNum, lines, i)
	}
}

func (s *goParseState) parseLine(line string, lineNum int, lines []string, idx int) {
	// Package
	if matches := goPackagePattern.FindStringSubmatch(line); len(matches) > 1 {
		s.ctx.Package = matches[1]
		return
	}

	// Handle import blocks
	if s.handleImports(line) {
		return
	}

	// Handle type blocks
	if s.handleTypes(line, lineNum) {
		return
	}

	// Handle functions
	if s.handleFunction(line, lineNum, lines, idx) {
		return
	}

	// Handle variables/constants
	s.handleVariable(line, lineNum)
}

func (s *goParseState) handleImports(line string) bool {
	if goImportBlockStart.MatchString(line) {
		s.inImportBlock = true
		return true
	}

	if s.inImportBlock {
		if strings.TrimSpace(line) == ")" {
			s.inImportBlock = false
			return true
		}
		if matches := goImportPattern.FindStringSubmatch(line); len(matches) > 2 {
			s.ctx.Imports = append(s.ctx.Imports, Import{Alias: matches[1], Path: matches[2]})
		}
		return true
	}

	// Single import
	if strings.HasPrefix(strings.TrimSpace(line), "import ") && !strings.Contains(line, "(") {
		if matches := goImportPattern.FindStringSubmatch(line); len(matches) > 2 {
			s.ctx.Imports = append(s.ctx.Imports, Import{Alias: matches[1], Path: matches[2]})
		}
		return true
	}

	return false
}

func (s *goParseState) handleTypes(line string, lineNum int) bool {
	if matches := goTypePattern.FindStringSubmatch(line); len(matches) > 2 {
		s.inTypeBlock = true
		s.currentType = matches[1]
		s.typeStartLine = lineNum
		s.braceCount = strings.Count(line, "{") - strings.Count(line, "}")
		return true
	}

	if s.inTypeBlock {
		s.braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		if s.braceCount <= 0 {
			s.addTypeDefinition(lineNum)
			s.inTypeBlock = false
			s.currentType = ""
		}
		return true
	}

	return false
}

func (s *goParseState) addTypeDefinition(lineNum int) {
	if strings.Contains(s.currentType, "interface") {
		s.ctx.Interfaces = append(s.ctx.Interfaces, Interface{
			Name: s.currentType, StartLine: s.typeStartLine, EndLine: lineNum, IsExported: isExported(s.currentType),
		})
	} else {
		s.ctx.Classes = append(s.ctx.Classes, Class{
			Name: s.currentType, StartLine: s.typeStartLine, EndLine: lineNum, IsExported: isExported(s.currentType),
		})
	}
}

func (s *goParseState) handleFunction(line string, lineNum int, lines []string, idx int) bool {
	matches := goFuncPattern.FindStringSubmatch(line)
	if len(matches) <= 2 {
		return false
	}

	fn := Function{
		Name: matches[2], Receiver: matches[1], StartLine: lineNum, IsExported: isExported(matches[2]),
	}

	if matches[3] != "" {
		fn.Parameters = parseGoParams(matches[3])
	}
	if matches[4] != "" {
		fn.Returns = parseGoReturns(matches[4])
	} else if matches[5] != "" {
		fn.Returns = []string{matches[5]}
	}

	fn.EndLine = findFunctionEnd(lines, idx) + 1
	s.ctx.Functions = append(s.ctx.Functions, fn)
	return true
}

func (s *goParseState) handleVariable(line string, lineNum int) {
	matches := goVarPattern.FindStringSubmatch(line)
	if len(matches) <= 1 {
		return
	}

	v := Variable{Name: matches[1], Type: matches[2], Line: lineNum, IsExported: isExported(matches[1])}
	if strings.HasPrefix(line, "const") {
		s.ctx.Constants = append(s.ctx.Constants, v)
	} else {
		s.ctx.Variables = append(s.ctx.Variables, v)
	}
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func parseGoParams(params string) []Param {
	var result []Param
	parts := strings.Split(params, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) >= 2 {
			result = append(result, Param{
				Name: fields[0],
				Type: strings.Join(fields[1:], " "),
			})
		} else if len(fields) == 1 {
			result = append(result, Param{Type: fields[0]})
		}
	}
	return result
}

func parseGoReturns(returns string) []string {
	var result []string
	parts := strings.Split(returns, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			// Extract just the type
			fields := strings.Fields(part)
			if len(fields) > 0 {
				result = append(result, fields[len(fields)-1])
			}
		}
	}
	return result
}

func findFunctionEnd(lines []string, startIdx int) int {
	braceCount := 0
	started := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if strings.Contains(line, "{") {
			started = true
		}

		if started && braceCount == 0 {
			return i
		}
	}
	return len(lines) - 1
}

// JavaScript/TypeScript parsing
func (p *Parser) parseJavaScript(lines []string, ctx *Context) {
	p.parseJSTS(lines, ctx)
}

func (p *Parser) parseTypeScript(lines []string, ctx *Context) {
	p.parseJSTS(lines, ctx)
}

func (p *Parser) parseJSTS(lines []string, ctx *Context) {
	importPattern := regexp.MustCompile(`^import\s+(?:{[^}]+}|[\w,\s]+)\s+from\s+['"]([^'"]+)['"]`)
	funcPattern := regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`)
	arrowPattern := regexp.MustCompile(`^(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*(?::\s*\w+)?\s*=>`)
	classPattern := regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`)

	for i, line := range lines {
		lineNum := i + 1

		// Imports
		if matches := importPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Imports = append(ctx.Imports, Import{Path: matches[1]})
			continue
		}

		// Functions
		if matches := funcPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Functions = append(ctx.Functions, Function{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "export"),
			})
			continue
		}

		// Arrow functions
		if matches := arrowPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Functions = append(ctx.Functions, Function{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "export"),
			})
			continue
		}

		// Classes
		if matches := classPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "export"),
			})
		}
	}
}

// Python parsing
func (p *Parser) parsePython(lines []string, ctx *Context) {
	importPattern := regexp.MustCompile(`^(?:from\s+(\S+)\s+)?import\s+(\S+)`)
	funcPattern := regexp.MustCompile(`^(?:async\s+)?def\s+(\w+)\s*\(`)
	classPattern := regexp.MustCompile(`^class\s+(\w+)`)

	for i, line := range lines {
		lineNum := i + 1

		// Imports
		if matches := importPattern.FindStringSubmatch(line); len(matches) > 1 {
			path := matches[1]
			if path == "" {
				path = matches[2]
			}
			ctx.Imports = append(ctx.Imports, Import{Path: path})
			continue
		}

		// Functions
		if matches := funcPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Functions = append(ctx.Functions, Function{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findPythonBlockEnd(lines, i) + 1,
				IsExported: !strings.HasPrefix(matches[1], "_"),
			})
			continue
		}

		// Classes
		if matches := classPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findPythonBlockEnd(lines, i) + 1,
				IsExported: !strings.HasPrefix(matches[1], "_"),
			})
		}
	}
}

func findPythonBlockEnd(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return startIdx
	}

	// Get indentation of the definition line
	defLine := lines[startIdx]
	defIndent := len(defLine) - len(strings.TrimLeft(defLine, " \t"))

	for i := startIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check indentation
		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
		if currentIndent <= defIndent {
			return i - 1
		}
	}

	return len(lines) - 1
}

// Java parsing
func (p *Parser) parseJava(lines []string, ctx *Context) {
	packagePattern := regexp.MustCompile(`^package\s+([\w.]+);`)
	importPattern := regexp.MustCompile(`^import\s+([\w.]+);`)
	classPattern := regexp.MustCompile(`^(?:public\s+)?(?:abstract\s+)?class\s+(\w+)`)
	interfacePattern := regexp.MustCompile(`^(?:public\s+)?interface\s+(\w+)`)
	methodPattern := regexp.MustCompile(`^(?:\s*)(?:public|private|protected)?\s*(?:static\s+)?(?:final\s+)?(\w+(?:<[^>]+>)?)\s+(\w+)\s*\(`)

	for i, line := range lines {
		lineNum := i + 1

		// Package
		if matches := packagePattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Package = matches[1]
			continue
		}

		// Imports
		if matches := importPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Imports = append(ctx.Imports, Import{Path: matches[1]})
			continue
		}

		// Classes
		if matches := classPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "public"),
			})
			continue
		}

		// Interfaces
		if matches := interfacePattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Interfaces = append(ctx.Interfaces, Interface{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "public"),
			})
			continue
		}

		// Methods
		if matches := methodPattern.FindStringSubmatch(line); len(matches) > 2 {
			ctx.Functions = append(ctx.Functions, Function{
				Name:       matches[2],
				Returns:    []string{matches[1]},
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "public"),
			})
		}
	}
}

// Rust parsing
func (p *Parser) parseRust(lines []string, ctx *Context) {
	usePattern := regexp.MustCompile(`^use\s+([\w:]+)`)
	fnPattern := regexp.MustCompile(`^(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`)
	structPattern := regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`)
	implPattern := regexp.MustCompile(`^impl(?:<[^>]+>)?\s+(\w+)`)
	traitPattern := regexp.MustCompile(`^(?:pub\s+)?trait\s+(\w+)`)

	for i, line := range lines {
		lineNum := i + 1

		// Use statements
		if matches := usePattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Imports = append(ctx.Imports, Import{Path: matches[1]})
			continue
		}

		// Functions
		if matches := fnPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Functions = append(ctx.Functions, Function{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "pub"),
			})
			continue
		}

		// Structs
		if matches := structPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "pub"),
			})
			continue
		}

		// Impl blocks
		if matches := implPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:       matches[1] + " (impl)",
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: true,
			})
			continue
		}

		// Traits
		if matches := traitPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Interfaces = append(ctx.Interfaces, Interface{
				Name:       matches[1],
				StartLine:  lineNum,
				EndLine:    findFunctionEnd(lines, i) + 1,
				IsExported: strings.Contains(line, "pub"),
			})
		}
	}
}

// Generic parsing for unsupported languages
func (p *Parser) parseGeneric(lines []string, ctx *Context) {
	funcPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^(?:func|function|def|fn|sub)\s+(\w+)`),
		regexp.MustCompile(`^(?:public|private|protected)?\s*(?:static\s+)?\w+\s+(\w+)\s*\(`),
	}
	classPattern := regexp.MustCompile(`^(?:class|struct|type)\s+(\w+)`)

	for i, line := range lines {
		lineNum := i + 1

		for _, pattern := range funcPatterns {
			if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
				ctx.Functions = append(ctx.Functions, Function{
					Name:      matches[1],
					StartLine: lineNum,
					EndLine:   findFunctionEnd(lines, i) + 1,
				})
				break
			}
		}

		if matches := classPattern.FindStringSubmatch(line); len(matches) > 1 {
			ctx.Classes = append(ctx.Classes, Class{
				Name:      matches[1],
				StartLine: lineNum,
				EndLine:   findFunctionEnd(lines, i) + 1,
			})
		}
	}
}
