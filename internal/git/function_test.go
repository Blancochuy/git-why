package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFunctionInContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		funcName  string
		parser    LanguageParser
		wantErr   bool
		wantStart int
		wantEnd   int
	}{
		{
			name: "find go function",
			content: `package main

func helper() int {
	return 1
}

func targetFunc(x int) int {
	result := x * 2
	return result
}

func another() {}`,
			funcName:  "targetFunc",
			parser:    languageParsers["go"],
			wantStart: 7,
			wantEnd:   10,
		},
		{
			name: "find go method",
			content: `type Service struct{}

func (s *Service) Process() error {
	return nil
}`,
			funcName:  "Process",
			parser:    languageParsers["go"],
			wantStart: 3,
			wantEnd:   5,
		},
		{
			name: "find ts function",
			content: `import { x } from 'y';

function helper() {
	return 1;
}

export async function fetchUser(id: string) {
	const user = await db.find(id);
	return user;
}

const other = () => {};`,
			funcName:  "fetchUser",
			parser:    languageParsers["ts"],
			wantStart: 7,
			wantEnd:   10,
		},
		{
			name: "find python function",
			content: `def helper():
    pass

def target_func(x):
    result = x * 2
    return result

def other():
    pass`,
			funcName:  "target_func",
			parser:    languageParsers["py"],
			wantStart: 4,
			wantEnd:   9,
		},
		{
			name: "function not found",
			content: `func foo() {}
func bar() {}`,
			funcName: "nonexistent",
			parser:   languageParsers["go"],
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findFunctionInContent(tt.content, tt.funcName, tt.parser)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStart, result.StartLine)
			assert.Equal(t, tt.wantEnd, result.EndLine)
			assert.Equal(t, tt.funcName, result.Name)
		})
	}
}

func TestCountBraces(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"open brace", "func foo() {", 1},
		{"close brace", "}", -1},
		{"multiple", "{{}}", 0},
		{"nested", "if x { if y { return } }", 0},
		{"string literal", `fmt.Println("{not a brace}")`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countBraces(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"main.go", "go"},
		{"auth.ts", "ts"},
		{"auth.tsx", "ts"},
		{"script.js", "js"},
		{"app.py", "py"},
		{"Main.java", "java"},
		{"lib.rs", "rs"},
		{"component.jsx", "ts"},
		{"noext", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getFileExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	langs := GetSupportedLanguages()
	assert.Contains(t, langs, "go")
	assert.Contains(t, langs, "ts")
	assert.Contains(t, langs, "js")
	assert.Contains(t, langs, "py")
	assert.Contains(t, langs, "java")
	assert.Contains(t, langs, "rs")
}
