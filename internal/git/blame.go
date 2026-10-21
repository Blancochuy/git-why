package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
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
	PRURL       string
	PRBody      string
}

type Repo struct {
	gitPath string
	Dir     string
}

func NewRepo() *Repo {
	return &Repo{gitPath: "git"}
}

func (r *Repo) runGit(args ...string) (string, error) {
	cmd := exec.Command(r.gitPath, args...)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
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

func (r *Repo) SetDirFromFile(file string) {
	if abs, err := filepath.Abs(file); err == nil {
		r.Dir = filepath.Dir(abs)
	}
}

func (r *Repo) GetRemoteURL(remoteName string) (string, error) {
	if remoteName == "" {
		remoteName = "origin"
	}
	output, err := r.runGit("config", "--get", "remote."+remoteName+".url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *Repo) Blame(file string, lineNum int) (*BlameResult, error) {
	rangeArg := fmt.Sprintf("-L%d,%d", lineNum, lineNum)
	// Add flags for Deep Blame: -w (ignore whitespace), -M (moves in same file), -C -C -C (copies across files)
	output, err := r.runGit("blame", "--porcelain", "-w", "-M", "-C", "-C", "-C", rangeArg, file)
	if err != nil {
		return nil, err
	}

	return parseBlameOutput(output)
}

func parseBlameOutput(output string) (*BlameResult, error) {
	result := &BlameResult{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		if result.Hash == "" {
			var hash string
			var orig, final, count int
			if _, err := fmt.Sscanf(line, "%s %d %d %d", &hash, &orig, &final, &count); err == nil && isHexHash(hash) {
				result.Hash = hash
				result.OriginalLineNum = orig
				result.FinalLineNum = final
				continue
			}
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

func IsUncommitted(hash string) bool {
	if hash == "" {
		return false
	}
	for _, c := range hash {
		if c != '0' {
			return false
		}
	}
	return true
}

func isHexHash(s string) bool {
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func parseTimestamp(s string) (time.Time, error) {
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}

func (r *Repo) SearchHistory(file, term string, isRegex bool) ([]CommitInfo, error) {
	// Unit separator between fields, record separator between commits, so
	// blank lines inside a commit body do not split the record.
	output, err := r.runGit("log", "--no-patch", "--format=%H%x1f%an%x1f%at%x1f%s%x1f%b%x1e", "--", file)
	if err != nil {
		return nil, err
	}

	return filterCommitsByTerm(output, term, isRegex)
}

func filterCommitsByTerm(output, term string, isRegex bool) ([]CommitInfo, error) {
	var commits []CommitInfo
	records := strings.Split(output, "\x1e")

	for _, record := range records {
		record = strings.Trim(record, "\r\n")
		if strings.TrimSpace(record) == "" {
			continue
		}

		parts := strings.SplitN(record, "\x1f", 5)
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

		if len(parts) == 5 {
			commit.Body = strings.Trim(parts[4], "\r\n")
		}

		matched, err := matchesTerm(commit, term, isRegex)
		if err != nil {
			return nil, err
		}
		if matched {
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

func matchesTerm(commit CommitInfo, term string, isRegex bool) (bool, error) {
	text := commit.Message + " " + commit.Body

	if isRegex {
		re, err := regexp.Compile(term)
		if err != nil {
			return false, fmt.Errorf("invalid search regex: %w", err)
		}
		return re.MatchString(text), nil
	}

	return strings.Contains(strings.ToLower(text), strings.ToLower(term)), nil
}
