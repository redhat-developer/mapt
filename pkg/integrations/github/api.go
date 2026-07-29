package github

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type registrationTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type runnersListResponse struct {
	Runners []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"runners"`
}

func newGHClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func ghRequest(method, url, pat string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+pat)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

// deleteRunnerByName lists runners at apiBase and deletes the one matching name.
// apiBase is the runners endpoint, e.g. https://api.github.com/orgs/{org}/actions/runners
func deleteRunnerByName(pat, apiBase, name string) error {
	req, err := ghRequest(http.MethodGet, apiBase, pat)
	if err != nil {
		return fmt.Errorf("creating list request: %w", err)
	}
	client := newGHClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("listing runners: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("listing runners: GitHub API returned %d: %s", resp.StatusCode, string(body))
	}
	var list runnersListResponse
	if err := json.Unmarshal(body, &list); err != nil {
		return fmt.Errorf("parsing runners list: %w", err)
	}
	for _, r := range list.Runners {
		if r.Name != name {
			continue
		}
		delURL := fmt.Sprintf("%s/%d", apiBase, r.ID)
		req, err := ghRequest(http.MethodDelete, delURL, pat)
		if err != nil {
			return fmt.Errorf("creating delete request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("deleting runner: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusNoContent {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("deleting runner: GitHub API returned %d: %s", resp.StatusCode, string(b))
		}
		return nil
	}
	return fmt.Errorf("runner %q not found", name)
}

// DeleteOrgRunner removes the named runner from the given GitHub organization.
func DeleteOrgRunner(pat, org, name string) error {
	return deleteRunnerByName(pat,
		fmt.Sprintf("https://api.github.com/orgs/%s/actions/runners", org),
		name)
}

// DeleteRepoRunner removes the named runner from the given repository.
// repoURL is "owner/repo" or "https://github.com/owner/repo".
func DeleteRepoRunner(pat, repoURL, name string) error {
	owner, repo, err := splitOwnerRepo(repoURL)
	if err != nil {
		return err
	}
	return deleteRunnerByName(pat,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners", owner, repo),
		name)
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

// GenerateOrgRegistrationToken calls the GitHub API to create a short-lived
// runner registration token for the given organization.
// pat is a Personal Access Token with admin:org scope.
func GenerateOrgRegistrationToken(pat, org string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/actions/runners/registration-token", org)

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
		return "", fmt.Errorf("GitHub API returned %d: %s (ensure GITHUB_TOKEN has admin:org scope)", resp.StatusCode, string(body))
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

// generateAppJWT creates a signed RS256 JWT for authenticating as a GitHub App.
func generateAppJWT(appID, privateKeyPath string) (string, error) {
	pemBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("reading private key %q: %w", privateKeyPath, err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", fmt.Errorf("no PEM block found in %q", privateKeyPath)
	}

	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = k
	} else if ki, err2 := x509.ParsePKCS8PrivateKey(block.Bytes); err2 == nil {
		var ok bool
		if key, ok = ki.(*rsa.PrivateKey); !ok {
			return "", fmt.Errorf("private key is not RSA")
		}
	} else {
		return "", fmt.Errorf("parsing private key: %w", err)
	}

	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(
		`{"iat":%d,"exp":%d,"iss":"%s"}`,
		now.Add(-60*time.Second).Unix(),
		now.Add(9*time.Minute).Unix(),
		appID,
	)))
	signingInput := header + "." + payload
	h := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// GenerateInstallationToken exchanges GitHub App credentials for a short-lived
// installation access token suitable for runner management API calls.
func GenerateInstallationToken(appID, installationID, privateKeyPath string) (string, error) {
	jwt, err := generateAppJWT(appID, privateKeyPath)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := newGHClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling GitHub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("empty token in installation token response")
	}
	return result.Token, nil
}
