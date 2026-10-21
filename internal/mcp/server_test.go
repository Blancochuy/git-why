package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLineRange(t *testing.T) {
	start, end, err := parseLineRange("10-20")
	require.NoError(t, err)
	assert.Equal(t, 10, start)
	assert.Equal(t, 20, end)

	_, _, err = parseLineRange("20-10")
	assert.Error(t, err)
	_, _, err = parseLineRange("nope")
	assert.Error(t, err)
}

func TestServerProtocol(t *testing.T) {
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`not-json`,
		`{"jsonrpc":"2.0","id":3,"method":"nope"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"missing"}}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	server.SetVersion("v0.1.0-test")
	require.NoError(t, server.Run())

	responses := decodeResponses(t, out.Bytes())
	require.Len(t, responses, 5)

	assert.Equal(t, "v0.1.0-test", responses[0].Result.ServerInfo["version"])
	assert.Equal(t, "git-why", responses[0].Result.ServerInfo["name"])

	var names []string
	for _, tool := range responses[1].Result.Tools {
		names = append(names, tool.Name)
	}
	assert.ElementsMatch(t, []string{"git-why-context", "git-why-stats"}, names)

	assert.Equal(t, -32700, responses[2].Error.Code)
	assert.Equal(t, -32601, responses[3].Error.Code)
	assert.Equal(t, -32602, responses[4].Error.Code)
}

func TestToolHandlersUseRepoHistory(t *testing.T) {
	dir := initGitRepo(t)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(""), &out)
	server.repo.Dir = dir

	server.callGitWhyContext(1, json.RawMessage(`{"file":"demo.go","line":3}`))
	server.callGitWhyContext(2, json.RawMessage(`{"file":"demo.go","range":"3-5"}`))
	server.callGitWhyStats(3, json.RawMessage(`{"file":"demo.go"}`))
	server.callGitWhyContext(4, json.RawMessage(`{"file":"demo.go"}`))

	responses := decodeResponses(t, out.Bytes())
	require.Len(t, responses, 4)

	lineText := responses[0].Result.Content[0].Text
	assert.Contains(t, lineText, "demo.go:3")
	assert.Contains(t, lineText, "feat: add targetFunc")
	assert.Contains(t, lineText, "Introduce the function")
	assert.False(t, responses[0].Result.IsError)

	rangeText := responses[1].Result.Content[0].Text
	assert.Contains(t, rangeText, "FILE: demo.go:3-5")
	assert.Contains(t, rangeText, "feat: add targetFunc")
	assert.False(t, responses[1].Result.IsError)

	statsText := responses[2].Result.Content[0].Text
	assert.Contains(t, statsText, "Stats for demo.go")
	assert.Contains(t, statsText, "Test User")
	assert.False(t, responses[2].Result.IsError)

	assert.True(t, responses[3].Result.IsError)
	assert.Contains(t, responses[3].Result.Content[0].Text, "line or a range")
}

type rpcResponse struct {
	Error *struct {
		Code int `json:"code"`
	} `json:"error"`
	Result struct {
		ServerInfo map[string]string `json:"serverInfo"`
		Tools      []struct {
			Name string `json:"name"`
		} `json:"tools"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
}

func decodeResponses(t *testing.T, raw []byte) []rpcResponse {
	t.Helper()
	var responses []rpcResponse
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var resp rpcResponse
		require.NoError(t, json.Unmarshal(line, &resp))
		responses = append(responses, resp)
	}
	return responses
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
