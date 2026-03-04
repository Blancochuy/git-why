package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/blancochuy/git-why/internal/enricher"
	"github.com/blancochuy/git-why/internal/git"
	"github.com/blancochuy/git-why/internal/render"
)

type Server struct {
	reader *bufio.Reader
	writer io.Writer
	repo   *git.Repo
}

func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		reader: bufio.NewReader(in),
		writer: out,
		repo:   git.NewRepo(),
	}
}

func (s *Server) Run() error {
	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error", nil)
			continue
		}

		s.handleRequest(&req)
	}
}

func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.sendResponse(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "git-why",
				"version": "0.1.0",
			},
		})

	case "notifications/initialized":
		// Ignore

	case "tools/list":
		s.handleListTools(req)

	case "tools/call":
		s.handleCallTool(req)

	default:
		s.sendError(req.ID, -32601, "Method not found", nil)
	}
}

func (s *Server) handleListTools(req *Request) {
	tools := []Tool{
		{
			Name:        "git-why-context",
			Description: "Explain the historical intent of a line or range of lines using Git context.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file": {
						Type:        "string",
						Description: "Path to the file (relative to the repository root)",
					},
					"line": {
						Type:        "integer",
						Description: "Specific line number (1-indexed)",
					},
					"range": {
						Type:        "string",
						Description: "Optional line range (for example: '10-20')",
					},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "git-why-stats",
			Description: "Show authorship and historical change statistics for a specific file.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file": {
						Type:        "string",
						Description: "Path to the file",
					},
				},
				Required: []string{"file"},
			},
		},
	}

	s.sendResponse(req.ID, ListToolsResult{Tools: tools})
}

func (s *Server) handleCallTool(req *Request) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", nil)
		return
	}

	switch params.Name {
	case "git-why-context":
		s.callGitWhyContext(req.ID, params.Arguments)
	case "git-why-stats":
		s.callGitWhyStats(req.ID, params.Arguments)
	default:
		s.sendError(req.ID, -32602, "Unknown tool", nil)
	}
}

func (s *Server) callGitWhyContext(id interface{}, args json.RawMessage) {
	var arguments struct {
		File  string `json:"file"`
		Line  int    `json:"line"`
		Range string `json:"range"`
	}
	if err := json.Unmarshal(args, &arguments); err != nil {
		s.sendError(id, -32602, "Invalid arguments", nil)
		return
	}

	if !s.repo.IsGitRepo() {
		s.repo.SetDirFromFile(arguments.File)
	}

	if !s.repo.IsGitRepo() {
		s.sendResponse(id, CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: "Error: not a Git repository for " + arguments.File}},
		})
		return
	}

	blame, err := s.repo.Blame(arguments.File, arguments.Line)
	if err != nil {
		s.sendResponse(id, CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error running git blame: %v", err)}},
		})
		return
	}

	commit, err := s.repo.ShowCommit(blame.Hash)
	if err != nil {
		s.sendResponse(id, CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error retrieving commit details: %v", err)}},
		})
		return
	}

	diff, _ := s.repo.GetCommitDiff(blame.Hash)
	commit.Diff = diff

	// Enrich with GitHub PR data if origin is GitHub
	remoteURL, err := s.repo.GetRemoteURL("origin")
	if err == nil && remoteURL != "" {
		owner, repoName, err := enricher.ParseGitHubRemote(remoteURL)
		if err == nil && owner != "" && repoName != "" {
			ghClient := enricher.NewGitHubClient("")
			if pr, _ := ghClient.GetPRForCommit(owner, repoName, commit.Hash); pr != nil {
				commit.PRURL = pr.HTMLURL
				commit.PRBody = pr.Body
			}
		}
	}

	renderer := render.NewAIRenderer()
	textOutput := renderer.GetContext(arguments.File, arguments.Line, blame, commit)

	s.sendResponse(id, CallToolResult{
		Content: []Content{{Type: "text", Text: textOutput}},
	})
}

func (s *Server) callGitWhyStats(id interface{}, args json.RawMessage) {
	var arguments struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(args, &arguments); err != nil {
		s.sendError(id, -32602, "Invalid arguments", nil)
		return
	}

	if !s.repo.IsGitRepo() {
		s.repo.SetDirFromFile(arguments.File)
	}

	stats, err := s.repo.Stats(arguments.File)
	if err != nil {
		s.sendResponse(id, CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error retrieving stats: %v", err)}},
		})
		return
	}

	output := fmt.Sprintf("Stats for %s\nTotal Lines: %d\nTotal Commits: %d\n\nContributors:\n",
		arguments.File, stats.TotalLines, stats.TotalCommits)

	for _, a := range stats.Authors {
		output += fmt.Sprintf("- %s: %d lines (%.1f%%)\n", a.Author, a.LinesCount, a.Percentage)
	}

	s.sendResponse(id, CallToolResult{
		Content: []Content{{Type: "text", Text: output}},
	})
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", data)
}

func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	dataBytes, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", dataBytes)
}
