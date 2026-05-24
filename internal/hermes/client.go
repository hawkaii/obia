package hermes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ContextBriefing is the session context returned by the Hermes API.
type ContextBriefing struct {
	Timestamp string            `json:"timestamp"`
	Repo      string            `json:"repo"`
	Briefing  string            `json:"briefing"`
	Recent    []RepoCommits     `json:"recent_commits"`
	VaultLogs []VaultLog        `json:"vault_logs"`
	DSA       DSAProgress       `json:"dsa"`
	Work      *WorkDigest       `json:"work_digest"`
	Unstaged  []UncommittedFile `json:"uncommitted"`
	Meta      struct {
		Ms int `json:"ms"`
	} `json:"meta"`
}

type RepoCommits struct {
	Repo    string   `json:"repo"`
	Commits []string `json:"commits"`
}

type VaultLog struct {
	File    string `json:"file"`
	Preview string `json:"preview"`
}

type DSAProgress struct {
	Solved   int              `json:"solved"`
	Total    int              `json:"total"`
	Problems []DSATask        `json:"problems"`
}

type DSATask struct {
	Num    int    `json:"num"`
	Name   string `json:"name"`
	Solved bool   `json:"solved"`
}

type WorkDigest struct {
	File    string `json:"file"`
	Content string `json:"content"`
}

type UncommittedFile struct {
	Status string `json:"status"`
	File   string `json:"file"`
}

type ChatRequest struct {
	Message string `json:"message"`
	Repo    string `json:"repo"`
}

type ChatResponse struct {
	Reply     string `json:"reply"`
	Timestamp string `json:"timestamp"`
}

// Client communicates with the Hermes Context API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Hermes API client.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:4123"
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchContext gets the session briefing from Hermes.
func (c *Client) FetchContext(repo string) (*ContextBriefing, error) {
	url := fmt.Sprintf("%s/context?repo=%s", c.BaseURL, repo)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("hermes request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var ctx ContextBriefing
	if err := json.Unmarshal(body, &ctx); err != nil {
		return nil, fmt.Errorf("parsing response: %w (body: %s)", err, string(body[:min(len(body), 200)]))
	}
	return &ctx, nil
}

// SendChat sends a message to the Hermes chat endpoint.
func (c *Client) SendChat(repo, message string) (*ChatResponse, error) {
	reqBody := ChatRequest{Message: message, Repo: repo}
	data, _ := json.Marshal(reqBody)

	resp, err := c.HTTPClient.Post(
		fmt.Sprintf("%s/chat", c.BaseURL),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("parsing chat response: %w", err)
	}
	return &chatResp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ChatResponseMsg is a Bubble Tea message for Hermes chat replies.
type ChatResponseMsg struct {
	Reply string
	Err   error
}
