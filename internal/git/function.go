package git

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type FunctionBounds struct {
	Name      string
	StartLine int
	EndLine   int
}

type LanguageParser struct {
	FunctionPatterns []*regexp.Regexp
	EndPattern       *regexp.Regexp
}

var languageParsers = map[string]LanguageParser{
	"go": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^func\s+(\w+)\s*\(`),
			regexp.MustCompile(`^func\s*\(\w+\s+\*?\w+\)\s+(\w+)\s*\(`),
		},
		EndPattern: regexp.MustCompile(`^\}`),
	},
	"ts": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(\s*)(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
			regexp.MustCompile(`^(\s*)(?:export\s+)?(?:public|private|protected)?\s*(?:async\s+)?(\w+)\s*\([^)]*\)\s*(?::\s*\w+)?\s*\{`),
			regexp.MustCompile(`^(\s*)(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[^=])*=>`),
		},
		EndPattern: regexp.MustCompile(`^\}`),
	},
	"js": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(\s*)(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
			regexp.MustCompile(`^(\s*)(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[^=])*=>`),
		},
		EndPattern: regexp.MustCompile(`^\}`),
	},
	"py": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(\s*)def\s+(\w+)\s*\(`),
			regexp.MustCompile(`^(\s*)async\s+def\s+(\w+)\s*\(`),
		},
		EndPattern: nil,
	},
	"java": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(\s*)(?:public|private|protected)?\s*(?:static\s+)?(?:\w+(?:<[^>]+>)?\s+)+(\w+)\s*\([^)]*\)\s*(?:throws\s+[\w,\s]+)?\s*\{`),
		},
		EndPattern: regexp.MustCompile(`^\s*\}`),
	},
	"rs": {
		FunctionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(\s*)(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`),
		},
		EndPattern: regexp.MustCompile(`^\}`),
	},
}

func (r *Repo) FindFunction(file, funcName string) (*FunctionBounds, error) {
	ext := getFileExtension(file)
	parser, ok := languageParsers[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", ext)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	return findFunctionInContent(string(content), funcName, parser)
}

func findFunctionInContent(content, funcName string, parser LanguageParser) (*FunctionBounds, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	var baseIndent string
	foundStart := false
	braceCount := 0
	bounds := &FunctionBounds{Name: funcName}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if !foundStart {
			for _, pattern := range parser.FunctionPatterns {
				matches := pattern.FindStringSubmatch(line)
				if len(matches) >= 2 {
					name := matches[len(matches)-1]
					if name == funcName {
						bounds.StartLine = lineNum
						foundStart = true
						if len(matches) > 2 {
							baseIndent = matches[1]
						}
						braceCount = countBraces(line)
						break
					}
				}
			}
			continue
		}

		braceCount += countBraces(line)

		if parser.EndPattern != nil {
			if braceCount <= 0 {
				bounds.EndLine = lineNum
				return bounds, nil
			}
		} else {
			if baseIndent != "" && !strings.HasPrefix(line, baseIndent) && strings.TrimSpace(line) != "" {
				bounds.EndLine = lineNum - 1
				return bounds, nil
			}
		}
	}

	if foundStart {
		bounds.EndLine = lineNum
		return bounds, nil
	}

	return nil, fmt.Errorf("function '%s' not found", funcName)
}

func countBraces(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case '{':
			count++
		case '}':
			count--
		}
	}
	return count
}

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}

	ext := parts[len(parts)-1]
	switch ext {
	case "tsx", "jsx":
		return "ts"
	case "py":
		return "py"
	case "java":
		return "java"
	case "rs":
		return "rs"
	case "go":
		return "go"
	case "ts":
		return "ts"
	case "js":
		return "js"
	default:
		return ext
	}
}

func GetSupportedLanguages() []string {
	langs := make([]string, 0, len(languageParsers))
	for lang := range languageParsers {
		langs = append(langs, lang)
	}
	return langs
}
