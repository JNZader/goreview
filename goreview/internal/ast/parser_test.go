package ast

import (
	"strings"
	"testing"
)

func TestParseGo(t *testing.T) {
	code := `package main

import (
	"fmt"
	"strings"
)

// Greeter greets people
type Greeter struct {
	name string
}

// Hello returns a greeting
func (g *Greeter) Hello() string {
	return fmt.Sprintf("Hello, %s!", g.name)
}

func main() {
	g := &Greeter{name: "World"}
	fmt.Println(g.Hello())
}
`

	parser := NewParser("go")
	ctx, err := parser.Parse(code, "main.go")

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if ctx.Package != "main" {
		t.Errorf("Expected package 'main', got '%s'", ctx.Package)
	}

	if len(ctx.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(ctx.Imports))
	}

	if len(ctx.Classes) != 1 {
		t.Errorf("Expected 1 struct, got %d", len(ctx.Classes))
	}

	if ctx.Classes[0].Name != "Greeter" {
		t.Errorf("Expected struct 'Greeter', got '%s'", ctx.Classes[0].Name)
	}

	if len(ctx.Functions) < 2 {
		t.Errorf("Expected at least 2 functions, got %d", len(ctx.Functions))
	}

	// Check for method with receiver
	foundMethod := false
	for _, fn := range ctx.Functions {
		if fn.Name == "Hello" && fn.Receiver != "" {
			foundMethod = true
			break
		}
	}
	if !foundMethod {
		t.Error("Expected to find Hello method with receiver")
	}
}

func TestParseJavaScript(t *testing.T) {
	code := `import { useState } from 'react';
import axios from 'axios';

export function fetchData(url) {
	return axios.get(url);
}

export const processData = async (data) => {
	return data.map(x => x * 2);
};

class DataService {
	constructor() {
		this.cache = {};
	}

	async fetch(id) {
		return this.cache[id];
	}
}
`

	parser := NewParser("javascript")
	ctx, err := parser.Parse(code, "service.js")

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(ctx.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(ctx.Imports))
	}

	if len(ctx.Functions) < 2 {
		t.Errorf("Expected at least 2 functions, got %d", len(ctx.Functions))
	}

	if len(ctx.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(ctx.Classes))
	}
}

func TestParsePython(t *testing.T) {
	code := `from typing import List, Optional
import json

def process_data(data: List[int]) -> List[int]:
    return [x * 2 for x in data]

class DataProcessor:
    def __init__(self):
        self.cache = {}

    def process(self, data):
        return self._internal_process(data)

    def _internal_process(self, data):
        return data
`

	parser := NewParser("python")
	ctx, err := parser.Parse(code, "processor.py")

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(ctx.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(ctx.Imports))
	}

	if len(ctx.Functions) < 1 {
		t.Errorf("Expected at least 1 function, got %d", len(ctx.Functions))
	}

	if len(ctx.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(ctx.Classes))
	}

	// Check export detection
	for _, fn := range ctx.Functions {
		if fn.Name == "process_data" && !fn.IsExported {
			t.Error("process_data should be exported (no underscore prefix)")
		}
	}
}

func TestParseRust(t *testing.T) {
	code := `use std::collections::HashMap;
use serde::{Deserialize, Serialize};

pub struct Config {
    name: String,
    value: i32,
}

impl Config {
    pub fn new(name: String) -> Self {
        Config { name, value: 0 }
    }

    pub fn get_name(&self) -> &str {
        &self.name
    }
}

pub trait Configurable {
    fn configure(&mut self);
}

fn internal_helper() -> bool {
    true
}
`

	parser := NewParser("rust")
	ctx, err := parser.Parse(code, "config.rs")

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(ctx.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(ctx.Imports))
	}

	if len(ctx.Classes) < 1 {
		t.Errorf("Expected at least 1 struct, got %d", len(ctx.Classes))
	}

	if len(ctx.Interfaces) < 1 {
		t.Errorf("Expected at least 1 trait, got %d", len(ctx.Interfaces))
	}

	// Check for private function
	for _, fn := range ctx.Functions {
		if fn.Name == "internal_helper" && fn.IsExported {
			t.Error("internal_helper should not be exported (no pub)")
		}
	}
}

func TestContextBuilder(t *testing.T) {
	code := `package main

func Hello() string {
	return "Hello"
}

func World() string {
	return "World"
}
`

	parser := NewParser("go")
	ctx, err := parser.Parse(code, "main.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	builder := NewContextBuilder(1000)
	result := builder.BuildPromptContext(ctx, nil)

	if !strings.Contains(result, "main.go") {
		t.Error("Context should contain file path")
	}

	if !strings.Contains(result, "Hello") {
		t.Error("Context should contain Hello function")
	}

	if !strings.Contains(result, "World") {
		t.Error("Context should contain World function")
	}
}

func TestBuildEnhancedRequest(t *testing.T) {
	fullContent := `package main

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}
`

	diff := `@@ -3,3 +3,3 @@ package main
 func Add(a, b int) int {
-	return a + b
+	return a + b + 1  // Bug fix
 }`

	builder := NewContextBuilder(2000)
	req, err := builder.BuildEnhancedRequest(diff, "go", "math.go", fullContent)

	if err != nil {
		t.Fatalf("BuildEnhancedRequest failed: %v", err)
	}

	if req.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", req.Language)
	}

	if req.FilePath != "math.go" {
		t.Errorf("Expected file path 'math.go', got '%s'", req.FilePath)
	}

	if req.StructuralCtx == "" {
		t.Error("Expected structural context to be populated")
	}
}

func TestIsExported(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"Hello", true},
		{"hello", false},
		{"HelloWorld", true},
		{"helloWorld", false},
		{"A", true},
		{"a", false},
		{"_private", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExported(tt.name); got != tt.expected {
				t.Errorf("isExported(%s) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func BenchmarkParseGo(b *testing.B) {
	code := strings.Repeat(`
func test() {
	x := 1
	return x
}
`, 100)

	parser := NewParser("go")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(code, "test.go")
	}
}
