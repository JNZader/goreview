package ast

import (
	"fmt"
	"strings"
)

// Format strings for consistent output (SonarQube S1192)
const (
	lineRangeFormat = " [lines %d-%d]\n"
)

// ContextBuilder builds enhanced context for LLM prompts
type ContextBuilder struct {
	maxContextLength int
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(maxLength int) *ContextBuilder {
	if maxLength <= 0 {
		maxLength = 2000
	}
	return &ContextBuilder{
		maxContextLength: maxLength,
	}
}

// BuildPromptContext creates a structured context string for LLM prompts
func (cb *ContextBuilder) BuildPromptContext(ctx *Context, dc *DiffContext) string {
	var sb strings.Builder

	cb.writeFileHeader(&sb, ctx)
	cb.writeImports(&sb, ctx.Imports)

	if dc != nil {
		cb.writeDiffContext(&sb, dc)
	} else {
		cb.writeFullContext(&sb, ctx)
	}

	return cb.truncateIfNeeded(sb.String())
}

func (cb *ContextBuilder) writeFileHeader(sb *strings.Builder, ctx *Context) {
	sb.WriteString(fmt.Sprintf("## File: %s\n", ctx.FilePath))
	sb.WriteString(fmt.Sprintf("Language: %s\n", ctx.Language))
	if ctx.Package != "" {
		sb.WriteString(fmt.Sprintf("Package: %s\n", ctx.Package))
	}
	sb.WriteString("\n")
}

func (cb *ContextBuilder) writeImports(sb *strings.Builder, imports []Import) {
	if len(imports) == 0 {
		return
	}

	sb.WriteString("### Dependencies:\n")
	for i, imp := range imports {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("... and %d more imports\n", len(imports)-10))
			break
		}
		if imp.Alias != "" {
			sb.WriteString(fmt.Sprintf("- %s as %s\n", imp.Path, imp.Alias))
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", imp.Path))
		}
	}
	sb.WriteString("\n")
}

func (cb *ContextBuilder) writeDiffContext(sb *strings.Builder, dc *DiffContext) {
	sb.WriteString("### Changed Functions:\n")
	for _, fn := range dc.ChangedFunctions {
		sb.WriteString(cb.formatFunction(fn))
	}

	if len(dc.ChangedClasses) > 0 {
		sb.WriteString("\n### Changed Classes/Structs:\n")
		for _, cls := range dc.ChangedClasses {
			sb.WriteString(cb.formatClass(cls))
		}
	}
}

func (cb *ContextBuilder) writeFullContext(sb *strings.Builder, ctx *Context) {
	if len(ctx.Functions) > 0 {
		sb.WriteString("### Functions:\n")
		for _, fn := range ctx.Functions {
			sb.WriteString(cb.formatFunction(fn))
		}
		sb.WriteString("\n")
	}

	if len(ctx.Classes) > 0 {
		sb.WriteString("### Classes/Structs:\n")
		for _, cls := range ctx.Classes {
			sb.WriteString(cb.formatClass(cls))
		}
		sb.WriteString("\n")
	}

	if len(ctx.Interfaces) > 0 {
		sb.WriteString("### Interfaces:\n")
		for _, iface := range ctx.Interfaces {
			sb.WriteString(cb.formatInterface(iface))
		}
		sb.WriteString("\n")
	}
}

func (cb *ContextBuilder) truncateIfNeeded(result string) string {
	if len(result) > cb.maxContextLength {
		return result[:cb.maxContextLength] + "\n... (truncated)"
	}
	return result
}

func (cb *ContextBuilder) formatFunction(fn Function) string {
	var sb strings.Builder

	visibility := "private"
	if fn.IsExported {
		visibility = "public"
	}

	// Function signature
	if fn.Receiver != "" {
		sb.WriteString(fmt.Sprintf("- (%s) %s.%s(", visibility, fn.Receiver, fn.Name))
	} else {
		sb.WriteString(fmt.Sprintf("- (%s) %s(", visibility, fn.Name))
	}

	// Parameters
	params := make([]string, len(fn.Parameters))
	for i, p := range fn.Parameters {
		if p.Name != "" {
			params[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
		} else {
			params[i] = p.Type
		}
	}
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(")")

	// Returns
	if len(fn.Returns) > 0 {
		sb.WriteString(" -> ")
		sb.WriteString(strings.Join(fn.Returns, ", "))
	}

	sb.WriteString(fmt.Sprintf(lineRangeFormat, fn.StartLine, fn.EndLine))

	return sb.String()
}

func (cb *ContextBuilder) formatClass(cls Class) string {
	var sb strings.Builder

	visibility := "private"
	if cls.IsExported {
		visibility = "public"
	}

	sb.WriteString(fmt.Sprintf("- (%s) %s", visibility, cls.Name))

	if cls.Extends != "" {
		sb.WriteString(fmt.Sprintf(" extends %s", cls.Extends))
	}

	if len(cls.Implements) > 0 {
		sb.WriteString(fmt.Sprintf(" implements %s", strings.Join(cls.Implements, ", ")))
	}

	sb.WriteString(fmt.Sprintf(lineRangeFormat, cls.StartLine, cls.EndLine))

	// Fields summary
	if len(cls.Fields) > 0 {
		sb.WriteString(fmt.Sprintf("  Fields: %d\n", len(cls.Fields)))
	}

	// Methods summary
	if len(cls.Methods) > 0 {
		sb.WriteString(fmt.Sprintf("  Methods: %s\n", strings.Join(cls.Methods, ", ")))
	}

	return sb.String()
}

func (cb *ContextBuilder) formatInterface(iface Interface) string {
	var sb strings.Builder

	visibility := "private"
	if iface.IsExported {
		visibility = "public"
	}

	sb.WriteString(fmt.Sprintf("- (%s) %s", visibility, iface.Name))
	sb.WriteString(fmt.Sprintf(lineRangeFormat, iface.StartLine, iface.EndLine))

	if len(iface.Methods) > 0 {
		sb.WriteString(fmt.Sprintf("  Methods: %s\n", strings.Join(iface.Methods, ", ")))
	}

	return sb.String()
}

// BuildCallGraph builds a simple call graph context
func (cb *ContextBuilder) BuildCallGraph(ctx *Context, targetFunction string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### Call context for %s:\n", targetFunction))

	// Find the target function
	var target *Function
	for i := range ctx.Functions {
		if ctx.Functions[i].Name == targetFunction {
			target = &ctx.Functions[i]
			break
		}
	}

	if target == nil {
		return fmt.Sprintf("Function %s not found in context\n", targetFunction)
	}

	// List other functions in the same file that might call or be called by this function
	sb.WriteString("Other functions in file:\n")
	for _, fn := range ctx.Functions {
		if fn.Name != targetFunction {
			sb.WriteString(fmt.Sprintf("  - %s [lines %d-%d]\n", fn.Name, fn.StartLine, fn.EndLine))
		}
	}

	return sb.String()
}

// EnhancedReviewRequest creates an enhanced review request with AST context
type EnhancedReviewRequest struct {
	Diff           string `json:"diff"`
	Language       string `json:"language"`
	FilePath       string `json:"file_path"`
	StructuralCtx  string `json:"structural_context"`
	ChangedSymbols string `json:"changed_symbols,omitempty"`
}

// BuildEnhancedRequest creates an enhanced review request
func (cb *ContextBuilder) BuildEnhancedRequest(
	diff, language, filePath, fullContent string,
) (*EnhancedReviewRequest, error) {
	parser := NewParser(language)

	// Parse full file
	ctx, err := parser.Parse(fullContent, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Parse diff context
	diffCtx, err := parser.ParseDiff(diff, fullContent, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	// Build structural context
	structuralCtx := cb.BuildPromptContext(ctx, diffCtx)

	// Build changed symbols summary
	var changedSymbols strings.Builder
	if len(diffCtx.ChangedFunctions) > 0 {
		changedSymbols.WriteString("Modified functions: ")
		names := make([]string, len(diffCtx.ChangedFunctions))
		for i, fn := range diffCtx.ChangedFunctions {
			names[i] = fn.Name
		}
		changedSymbols.WriteString(strings.Join(names, ", "))
	}
	if len(diffCtx.ChangedClasses) > 0 {
		if changedSymbols.Len() > 0 {
			changedSymbols.WriteString("; ")
		}
		changedSymbols.WriteString("Modified classes: ")
		names := make([]string, len(diffCtx.ChangedClasses))
		for i, cls := range diffCtx.ChangedClasses {
			names[i] = cls.Name
		}
		changedSymbols.WriteString(strings.Join(names, ", "))
	}

	return &EnhancedReviewRequest{
		Diff:           diff,
		Language:       language,
		FilePath:       filePath,
		StructuralCtx:  structuralCtx,
		ChangedSymbols: changedSymbols.String(),
	}, nil
}
