package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"
)

func main() {
	// Parse inputs
	input := &Input{
		Repository:    getInput("repository", ""),
		GithubToken:   getInput("github_token", ""),
		Branch:        getInput("branch", "main"),
		ValuesFile:    getInput("values_file", "app/values/values.beta.yaml"),
		VersionKey:    getInput("version_key", "version"),
		Version:       getInput("version", ""),
		CommitMessage: getInput("commit_message", "update version"),
	}

	// Validate inputs
	if err := validateInput(input); err != nil {
		fmt.Printf("::error::Input validation failed: %v\n", err)
		os.Exit(1)
	}

	// Get current file content
	fileResp, err := getFileContent(input)
	if err != nil {
		fmt.Printf("::error::Failed to get file content: %v\n", err)
		os.Exit(1)
	}

	// Decode content
	content, err := base64.StdEncoding.DecodeString(fileResp.Content)
	if err != nil {
		fmt.Printf("::error::Failed to decode file content: %v\n", err)
		os.Exit(1)
	}

	// Update version in content
	newContent, err := updateVersionInContent(string(content), input.VersionKey, input.Version)
	if err != nil {
		fmt.Printf("::error::Failed to update version in content: %v\n", err)
		os.Exit(1)
	}

	// Update file in repository
	if err := updateFile(input, fileResp, newContent); err != nil {
		fmt.Printf("::error::Failed to update file in repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("::success::Successfully updated version in YAML file")
}

// Input represents the action inputs
type Input struct {
	Repository    string
	GithubToken   string
	Branch        string
	ValuesFile    string
	VersionKey    string
	Version       string
	CommitMessage string
}

// GitHubFileResponse represents the GitHub API response for getting a file
type GitHubFileResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	SHA      string `json:"sha"`
}

// GitHubUpdateRequest represents the request body for updating a file
type GitHubUpdateRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
	Branch  string `json:"branch"`
}

func getInput(key, defaultValue string) string {
	value := os.Getenv("INPUT_" + strings.ToUpper(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func validateInput(input *Input) error {
	if input.Repository == "" {
		return fmt.Errorf("repository input is required")
	}
	if input.GithubToken == "" {
		return fmt.Errorf("github_token input is required")
	}
	return nil
}

func getFileContent(input *Input) (*GitHubFileResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos%s/contents/%s?ref=%s",
		input.Repository, input.ValuesFile, input.Branch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "token "+input.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get file (status %d): %s", resp.StatusCode, string(body))
	}

	var fileResp GitHubFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &fileResp, nil
}

func updateFile(input *Input, fileResp *GitHubFileResponse, newContent string) error {
	url := fmt.Sprintf("https://api.github.com/repos%s/contents/%s",
		input.Repository, input.ValuesFile)

	if input.CommitMessage == "" {
		input.CommitMessage = fmt.Sprintf("chore: update %s to %s", input.VersionKey,
			input.Version)
	}
	updateReq := GitHubUpdateRequest{
		Message: input.CommitMessage,
		Content: base64.StdEncoding.EncodeToString([]byte(newContent)),
		SHA:     fileResp.SHA,
		Branch:  input.Branch,
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %v", err)
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create update request: %v", err)
	}

	req.Header.Set("Authorization", "token "+input.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update file (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func updateVersionInContent(src, versionKey, value string) (string, error) {
	keyPath := strings.Split(versionKey, ".")
	position, key, ok := findYAMLKeyPosition(src, keyPath)
	if !ok {
		return "", fmt.Errorf("key %s not found", strings.Join(keyPath, "/"))
	}
	keyStart := strings.Index(src[position:], key)
	keyEnd := keyStart + len(key)
	keyLineEnd := strings.Index(src[position+keyEnd:], "\n")
	valueSet := value
	if unicode.IsDigit(rune(value[0])) {
		valueSet = `"` + value + `"`
	}
	return src[0:position+keyEnd] + ": " + valueSet + src[position+keyEnd+keyLineEnd:], nil
}

func findYAMLKeyPosition(yamlData string, keyPath []string) (pos int,
	key string, found bool) {
	if len(keyPath) == 0 {
		return 0, "", false
	}
	lines := strings.Split(yamlData, "\n")
	indent := 0
	re := regexp.MustCompile(`^(\s*)`)
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		curIndent := len(re.FindString(line)) / 2
		if curIndent < indent {
			indent = curIndent
		}
		if curIndent == indent {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key = strings.TrimSpace(parts[0])
			if key == keyPath[0] {
				if len(keyPath) == 1 {
					return pos + strings.Index(line, key), key, true
				}
				keyPath = keyPath[1:]
				indent++
			}
		}
		pos += len(line) + 1
	}
	return 0, "", false
}
