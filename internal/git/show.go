package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (r *Repo) Show(file string, lineNum, context int) (string, error) {
	start := lineNum - context
	if start < 1 {
		start = 1
	}
	end := lineNum + context

	rangeArg := fmt.Sprintf("-L%d,%d:%s", start, end, file)

	cmd := exec.Command("git", "log", "-p", "--no-patch", rangeArg, file)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

func (r *Repo) GetContextLines(file string, lineNum, context int) ([]string, error) {
	start := lineNum - context
	if start < 1 {
		start = 1
	}
	end := lineNum + context

	output, err := r.runGit("blame", "-L", fmt.Sprintf("%d,%d", start, end), file)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func (r *Repo) GetCommitDiff(hash string) (string, error) {
	output, err := r.runGit("show", "--no-notes", "--format=", hash)
	if err != nil {
		return "", err
	}
	return output, nil
}
