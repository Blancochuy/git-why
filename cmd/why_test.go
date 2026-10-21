package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTarget(t *testing.T) {
	file, line, end, err := parseTarget("internal/git/blame.go:93")
	require.NoError(t, err)
	assert.Equal(t, "internal/git/blame.go", file)
	assert.Equal(t, 93, line)
	assert.Equal(t, 0, end)

	file, line, end, err = parseTarget("internal/git/blame.go:93-110")
	require.NoError(t, err)
	assert.Equal(t, "internal/git/blame.go", file)
	assert.Equal(t, 93, line)
	assert.Equal(t, 110, end)

	_, _, _, err = parseTarget("internal/git/blame.go:20-10")
	assert.Error(t, err)

	_, _, _, err = parseTarget("internal/git/blame.go")
	assert.Error(t, err)
}

func TestRunWhyLineRangeAndFunction(t *testing.T) {
	dir := initGitRepo(t)
	restore := chdir(t, dir)
	defer restore()

	lineOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"demo.go:3", "--plain"})
		require.NoError(t, rootCmd.Execute())
	})
	assert.Contains(t, lineOut, "Line 3")
	assert.Contains(t, lineOut, "feat: add targetFunc")
	assert.Contains(t, lineOut, "Introduce the function")
	assert.NotContains(t, lineOut, "\033[")

	shortOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"demo.go:3", "--short"})
		require.NoError(t, rootCmd.Execute())
	})
	assert.Contains(t, shortOut, "feat: add targetFunc")
	assert.Contains(t, shortOut, "@Test User")

	var payload OutputJSON
	jsonOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"demo.go:3", "--json"})
		require.NoError(t, rootCmd.Execute())
	})
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	assert.Equal(t, "demo.go", payload.File)
	assert.Equal(t, 3, payload.Line)
	assert.Equal(t, "feat: add targetFunc", payload.Summary)
	assert.Contains(t, payload.Body, "Introduce the function")

	var rangePayload OutputJSONRange
	rangeOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"demo.go:3-5", "--json"})
		require.NoError(t, rootCmd.Execute())
	})
	require.NoError(t, json.Unmarshal([]byte(rangeOut), &rangePayload))
	assert.Equal(t, "3-5", rangePayload.Range)
	require.NotEmpty(t, rangePayload.Commits)
	assert.Equal(t, "feat: add targetFunc", rangePayload.Commits[0].Message)

	fnOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"demo.go", "--fn", "targetFunc", "--plain"})
		require.NoError(t, rootCmd.Execute())
	})
	assert.Contains(t, fnOut, "Function targetFunc() in demo.go (lines 3-5)")
	assert.Contains(t, fnOut, "feat: add targetFunc")

	whyOut := captureStdout(t, func() {
		resetWhyFlags()
		rootCmd.SetArgs([]string{"why", "demo.go", "--fn", "targetFunc", "--plain"})
		require.NoError(t, rootCmd.Execute())
	})
	assert.Contains(t, whyOut, "Function targetFunc() in demo.go (lines 3-5)")

	resetWhyFlags()
	rootCmd.SetArgs([]string{"demo.go"})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid target format")
}

func TestUncommittedLine(t *testing.T) {
	dir := initGitRepo(t)
	restore := chdir(t, dir)
	defer restore()

	f, err := os.OpenFile("demo.go", os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("\nfunc extra() {}\n")
	require.NoError(t, f.Close())
	require.NoError(t, err)

	resetWhyFlags()
	rootCmd.SetArgs([]string{"demo.go:7", "--plain"})
	err = rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uncommitted")
}

func resetWhyFlags() {
	set := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	set(rootCmd.PersistentFlags().Set("short", "false"))
	set(rootCmd.PersistentFlags().Set("plain", "false"))
	set(rootCmd.PersistentFlags().Set("json", "false"))
	set(rootCmd.PersistentFlags().Set("md", "false"))
	set(rootCmd.PersistentFlags().Set("llm", "false"))
	set(rootCmd.PersistentFlags().Set("include-diff", "false"))
	set(rootCmd.Flags().Set("fn", ""))
	set(whyCmd.Flags().Set("fn", ""))
	funcName = ""
	shortOutput = false
	plainOutput = false
	jsonOutput = false
	mdOutput = false
	llmOutput = false
	includeDiff = false
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)
	return buf.String()
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	return func() {
		resetWhyFlags()
		rootCmd.SetArgs(nil)
		require.NoError(t, os.Chdir(wd))
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test User",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test User",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s: %s", strings.Join(args, " "), out)
	}

	git("init", "-b", "main")
	content := "package demo\n\nfunc targetFunc() int {\n\treturn 1\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "demo.go"), []byte(content), 0o644))
	git("add", "demo.go")
	git("commit", "-m", "feat: add targetFunc\n\nIntroduce the function so callers can see why it exists.")
	return dir
}
