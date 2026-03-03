package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type BlameResult struct {
	Hash            string
	Author          string
	AuthorMail      string
	Date            time.Time
	Line            string
	OriginalLineNum int
	FinalLineNum    int
}

type CommitInfo struct {
	Hash        string
	Author      string
	AuthorEmail string
	Date        time.Time
	Message     string
	Body        string
	Diff        string
}

type Repo struct {
	gitPath string
}

func NewRepo() *Repo {
	return &Repo{gitPath: "git"}
}

func (r *Repo) runGit(args ...string) (string, error) {
	cmd := exec.Command(r.gitPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), stderr.String())
	}
	return stdout.String(), nil
}

func (r *Repo) IsGitRepo() bool {
	_, err := r.runGit("rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (r *Repo) Blame(file string, lineNum int) (*BlameResult, error) {
	rangeArg := fmt.Sprintf("-L%d,%d", lineNum, lineNum)
	output, err := r.runGit("blame", "--porcelain", rangeArg, file)
	if err != nil {
		return nil, err
	}

	return parseBlameOutput(output)
}

func parseBlameOutput(output string) (*BlameResult, error) {
	result := &BlameResult{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	hashPattern := regexp.MustCompile(`^([a-f0-9]+)\s`)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := hashPattern.FindStringSubmatch(line); matches != nil && result.Hash == "" {
			result.Hash = matches[1]
		}

		if strings.HasPrefix(line, "author ") {
			result.Author = strings.TrimSpace(strings.TrimPrefix(line, "author "))
		}
		if strings.HasPrefix(line, "author-mail ") {
			result.AuthorMail = strings.TrimSpace(strings.TrimPrefix(line, "author-mail "))
		}
		if strings.HasPrefix(line, "author-time ") {
			ts := strings.TrimSpace(strings.TrimPrefix(line, "author-time "))
			if t, err := parseTimestamp(ts); err == nil {
				result.Date = t
			}
		}
		if strings.HasPrefix(line, "\t") {
			result.Line = strings.TrimPrefix(line, "\t")
		}
	}

	if result.Hash == "" {
		return nil, fmt.Errorf("could not parse blame output")
	}

	return result, nil
}

func (r *Repo) ShowCommit(hash string) (*CommitInfo, error) {
	output, err := r.runGit("show", "--no-patch", "--format=%H%n%an%n%ae%n%at%n%s%n%b", hash)
	if err != nil {
		return nil, err
	}

	return parseCommitInfo(output)
}

func parseCommitInfo(output string) (*CommitInfo, error) {
	output = strings.ReplaceAll(output, "\r", "")
	lines := strings.SplitN(output, "\n", 6)
	if len(lines) < 5 {
		return nil, fmt.Errorf("invalid commit info format")
	}

	result := &CommitInfo{
		Hash:        lines[0],
		Author:      lines[1],
		AuthorEmail: lines[2],
	}

	if ts := lines[3]; ts != "" {
		if t, err := parseTimestamp(ts); err == nil {
			result.Date = t
		}
	}

	result.Message = lines[4]
	if len(lines) > 5 {
		result.Body = lines[5]
	}

	return result, nil
}

func (r *Repo) LogRange(file string, start, end int) ([]CommitInfo, error) {
	rangeArg := fmt.Sprintf("-L%d,%d:%s", start, end, file)
	output, err := r.runGit("log", "--no-patch", "--format=%H|%an|%at|%s", rangeArg)
	if err != nil {
		return nil, err
	}

	return parseLogOutput(output)
}

func parseLogOutput(output string) ([]CommitInfo, error) {
	var commits []CommitInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}

		commit := CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Message: parts[3],
		}

		if t, err := parseTimestamp(parts[2]); err == nil {
			commit.Date = t
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func parseTimestamp(s string) (time.Time, error) {
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}

func (r *Repo) SearchHistory(file, term string, isRegex bool) ([]CommitInfo, error) {
	output, err := r.runGit("log", "--no-patch", "--format=%H|%an|%at|%s%n%b", "--", file)
	if err != nil {
		return nil, err
	}

	return filterCommitsByTerm(output, term, isRegex)
}

func filterCommitsByTerm(output, term string, isRegex bool) ([]CommitInfo, error) {
	var commits []CommitInfo
	blocks := strings.Split(output, "\n\n")

	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 1 {
			continue
		}

		parts := strings.SplitN(lines[0], "|", 4)
		if len(parts) < 4 {
			continue
		}

		commit := CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Message: parts[3],
		}

		if t, err := parseTimestamp(parts[2]); err == nil {
			commit.Date = t
		}

		if len(lines) > 1 {
			commit.Body = strings.Join(lines[1:], "\n")
		}

		if matchesTerm(commit, term, isRegex) {
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

func matchesTerm(commit CommitInfo, term string, isRegex bool) bool {
	text := commit.Message + " " + commit.Body

	if isRegex {
		matched, _ := regexp.MatchString(term, text)
		return matched
	}

	return strings.Contains(strings.ToLower(text), strings.ToLower(term))
}
