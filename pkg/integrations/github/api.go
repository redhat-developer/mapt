package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type registrationTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// splitOwnerRepo parses a repo URL in the form "owner/repo" or
// "https://github.com/owner/repo" and returns the two components.
func splitOwnerRepo(repoURL string) (owner, repo string, err error) {
	ownerRepo := strings.TrimPrefix(repoURL, "https://github.com/")
	ownerRepo = strings.TrimPrefix(ownerRepo, "http://github.com/")
	ownerRepo = strings.TrimSuffix(ownerRepo, "/")

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo format %q, expected owner/repo", repoURL)
	}
	return parts[0], parts[1], nil
}

// GenerateRegistrationToken calls the GitHub API to create a short-lived
// runner registration token for the given repository.
// pat is a Personal Access Token with repo admin scope.
// repoURL is in the form "owner/repo" or "https://github.com/owner/repo".
func GenerateRegistrationToken(pat, repoURL string) (string, error) {
	owner, repo, err := splitOwnerRepo(repoURL)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners/registration-token", owner, repo)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+pat)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling GitHub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned %d: %s (ensure GITHUB_TOKEN has admin scope on the repo)", resp.StatusCode, string(body))
	}

	var tokenResp registrationTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if tokenResp.Token == "" {
		return "", fmt.Errorf("empty token in GitHub API response")
	}

	return tokenResp.Token, nil
}
