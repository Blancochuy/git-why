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

type GitLabClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
	cache      *cache.Cache
}

type MergeRequest struct {
	IID          int       `json:"iid"`
	Title        string    `json:"title"`
	State        string    `json:"state"`
	WebURL       string    `json:"web_url"`
	Author       Author    `json:"author"`
	MergedAt     time.Time `json:"merged_at"`
	CreatedAt    time.Time `json:"created_at"`
	Labels       []string  `json:"labels"`
	TargetBranch string    `json:"target_branch"`
}

type Author struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

func NewGitLabClient(token, baseURL string) *GitLabClient {
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
		if token == "" {
			token = os.Getenv("GITWHY_GITLAB_TOKEN")
		}
	}

	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	return &GitLabClient{
		token:   token,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: cache.New(),
	}
}

func (c *GitLabClient) GetMRForCommit(projectPath, commitSHA string) (*MergeRequest, error) {
	cacheKey := fmt.Sprintf("mr:%s:%s", projectPath, commitSHA)
	if cached, ok := c.cache.Get(cacheKey); ok {
		var mr MergeRequest
		if err := json.Unmarshal(cached, &mr); err == nil {
			return &mr, nil
		}
	}

	encodedProject := url.PathEscape(projectPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s/merge_requests", c.baseURL, encodedProject, commitSHA)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
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
		return nil, fmt.Errorf("gitlab api error: %d", resp.StatusCode)
	}

	var mrs []MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mrs); err != nil {
		return nil, err
	}

	if len(mrs) == 0 {
		return nil, nil
	}

	mr := &mrs[0]
	if data, err := json.Marshal(mr); err == nil {
		c.cache.Set(cacheKey, data, 24*time.Hour)
	}

	return mr, nil
}

func (c *GitLabClient) GetIssue(projectPath string, iid int) (*GitLabIssue, error) {
	cacheKey := fmt.Sprintf("gl_issue:%s:%d", projectPath, iid)
	if cached, ok := c.cache.Get(cacheKey); ok {
		var issue GitLabIssue
		if err := json.Unmarshal(cached, &issue); err == nil {
			return &issue, nil
		}
	}

	encodedProject := url.PathEscape(projectPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/issues/%d", c.baseURL, encodedProject, iid)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
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
		return nil, fmt.Errorf("gitlab api error: %d", resp.StatusCode)
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	if data, err := json.Marshal(issue); err == nil {
		c.cache.Set(cacheKey, data, 24*time.Hour)
	}

	return &issue, nil
}

type GitLabIssue struct {
	IID       int       `json:"iid"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	WebURL    string    `json:"web_url"`
	Author    Author    `json:"author"`
	Labels    []string  `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
}

func ParseGitLabRemote(remoteURL string) (projectPath string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.TrimSuffix(remoteURL, "/")

	if strings.HasPrefix(remoteURL, "git@gitlab.com:") {
		return strings.TrimPrefix(remoteURL, "git@gitlab.com:"), nil
	}

	if strings.Contains(remoteURL, "gitlab.com") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", err
		}
		return strings.Trim(u.Path, "/"), nil
	}

	return "", fmt.Errorf("not a gitlab remote")
}

func (c *GitLabClient) ExtractIssueNumbers(message string) []int {
	var numbers []int
	seen := make(map[int]bool)

	pattern := regexp.MustCompile(`(?i)(?:fixes|closes|resolves|references|refs|issue)\s+#(\d+)`)
	matches := pattern.FindAllStringSubmatch(message, -1)

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
