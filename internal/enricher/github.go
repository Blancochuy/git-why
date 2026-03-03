package enricher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chuy/git-why/internal/cache"
)

type GitHubClient struct {
	token      string
	httpClient *http.Client
	cache      *cache.Cache
}

type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	MergedAt  time.Time `json:"merged_at"`
	CreatedAt time.Time `json:"created_at"`
	Labels    []Label   `json:"labels"`
}

type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	Labels    []Label   `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	Login string `json:"login"`
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PRSearchResponse struct {
	Items []PullRequest `json:"items"`
}

func NewGitHubClient(token string) *GitHubClient {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITWHY_TOKEN")
		}
	}

	return &GitHubClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: cache.New(),
	}
}

func (c *GitHubClient) GetPRForCommit(owner, repo, commitSHA string) (*PullRequest, error) {
	cacheKey := fmt.Sprintf("pr:%s:%s:%s", owner, repo, commitSHA)
	if cached, ok := c.cache.Get(cacheKey); ok {
		var pr PullRequest
		if err := json.Unmarshal(cached, &pr); err == nil {
			return &pr, nil
		}
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s/pulls", owner, repo, commitSHA)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		return nil, nil
	}

	pr := &prs[0]
	if data, err := json.Marshal(pr); err == nil {
		c.cache.Set(cacheKey, data, 24*time.Hour)
	}

	return pr, nil
}

func (c *GitHubClient) GetIssue(owner, repo string, number int) (*Issue, error) {
	cacheKey := fmt.Sprintf("issue:%s:%s:%d", owner, repo, number)
	if cached, ok := c.cache.Get(cacheKey); ok {
		var issue Issue
		if err := json.Unmarshal(cached, &issue); err == nil {
			return &issue, nil
		}
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	if data, err := json.Marshal(issue); err == nil {
		c.cache.Set(cacheKey, data, 24*time.Hour)
	}

	return &issue, nil
}

func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.TrimSuffix(remoteURL, "/")

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	if strings.HasPrefix(remoteURL, "https://github.com/") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", "", err
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("not a github remote")
}

func (c *GitHubClient) ExtractIssueNumbers(message string) []int {
	var numbers []int
	seen := make(map[int]bool)

	fixesPattern := regexp.MustCompile(`(?i)(?:fixes|closes|resolves|references|refs|issue)\s+#(\d+)`)
	matches := fixesPattern.FindAllStringSubmatch(message, -1)
	for _, match := range matches {
		if len(match) > 1 {
			var num int
			if _, err := fmt.Sscanf(match[1], "%d", &num); err == nil && !seen[num] {
				seen[num] = true
				numbers = append(numbers, num)
			}
		}
	}

	return numbers
}
