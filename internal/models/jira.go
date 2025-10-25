package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string          `json:"summary"`
		Description json.RawMessage `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

type JiraSearchResult struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type JiraProjectResult struct {
	Values []JiraProject `json:"values"`
}

type JiraIssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type JiraCreateMetaResponse struct {
	Projects []struct {
		Key        string          `json:"key"`
		IssueTypes []JiraIssueType `json:"issuetypes"`
	} `json:"projects"`
}

func FetchAssignedIssues(domain, email, apiToken string) ([]JiraIssue, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/search/jql?jql=assignee=currentUser()&fields=summary,issuetype,key,description", domain)
	req, _ := http.NewRequest("GET", url, nil)

	decryptedToken := tryDecrypt(apiToken)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Jira API error (%d): %s", res.StatusCode, string(body))
	}

	var result JiraSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Issues, nil
}

func ExtractDescriptionText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "(Keine Beschreibung vorhanden)"
	}

	var adf struct {
		Content []ADFNode `json:"content"`
	}
	if err := json.Unmarshal(raw, &adf); err != nil {
		return "(Fehler beim Lesen der Beschreibung)"
	}

	text := extractTextRecursive(adf.Content, 0)
	if text == "" {
		text = "(Keine Beschreibung vorhanden)"
	}
	return text
}

type ADFNode struct {
	Type    string      `json:"type"`
	Text    string      `json:"text,omitempty"`
	Content []ADFNode   `json:"content,omitempty"`
	Items   [][]ADFNode `json:"items,omitempty"`
}

// unterstützt Absätze, Bullet- & Ordered-Lists
func extractTextRecursive(nodes []ADFNode, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	for _, n := range nodes {
		switch n.Type {

		case "paragraph":
			sb.WriteString(extractTextRecursive(n.Content, indent))
			sb.WriteString("\n") // nur ein einfacher Zeilenumbruch

		case "text":
			sb.WriteString(n.Text)

		case "bulletList", "orderedList":
			listItems := n.Items
			if len(listItems) == 0 && len(n.Content) > 0 {
				listItems = [][]ADFNode{n.Content}
			}

			for i, item := range listItems {
				marker := "•"
				if n.Type == "orderedList" {
					marker = fmt.Sprintf("%d.", i+1)
				}
				sb.WriteString(fmt.Sprintf("%s%s %s\n",
					prefix,
					marker,
					strings.TrimSpace(extractTextRecursive(item, indent+1)),
				))
			}

		case "listItem":
			// nur Inhalte rendern, kein zusätzliches Bullet hier
			sb.WriteString(extractTextRecursive(n.Content, indent+1))

		default:
			sb.WriteString(extractTextRecursive(n.Content, indent))
		}
	}

	return sb.String()
}

func FetchFavouriteProjects(domain, email, token string) ([]JiraProject, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/project/search?favourite=true", domain)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := tryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Jira API error (%d): %s", res.StatusCode, string(body))
	}

	var result JiraProjectResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Values, nil
}

func CreateJiraIssue(domain, email, token, projectKey, issueType, title, content string, labels []string) error {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue", domain)

	// Beschreibung im Atlassian Document Format (ADF)
	adf := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": content,
					},
				},
			},
		},
	}

	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"issuetype": map[string]string{
				"name": issueType,
			},
			"summary":     title,
			"description": adf,
			"labels":      labels,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	decryptedToken := tryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create issue: %s", string(b))
	}

	return nil
}

func FetchProjectIssueTypes(domain, email, token, projectKey string) ([]JiraIssueType, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue/createmeta?projectKeys=%s", domain, projectKey)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := tryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Jira API error (%d): %s", res.StatusCode, string(body))
	}

	var data JiraCreateMetaResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Projects) == 0 {
		return nil, fmt.Errorf("keine Projektmetadaten für %s gefunden", projectKey)
	}

	return data.Projects[0].IssueTypes, nil
}

type jiraSearchResult struct {
	Issues []struct {
		Fields struct {
			Labels []string `json:"labels"`
		} `json:"fields"`
	} `json:"issues"`
}

// FetchAllProjects returns all visible projects (first page) for the user
func FetchAllProjects(domain, email, token string) ([]JiraProject, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/project/search?favourite=true", domain)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := tryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Jira API error (%d): %s", res.StatusCode, string(b))
	}

	var out JiraProjectResult
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Values, nil
}

// FetchProjectLabels queries all issues in a project and aggregates unique labels
func FetchProjectLabels(domain, email, token, projectKey string) ([]string, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/search/jql", domain)

	body := map[string]interface{}{
		"jql":        fmt.Sprintf("project=%s", projectKey),
		"fields":     []string{"labels"},
		"maxResults": 1000,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	decryptedToken := tryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Jira API error (%d): %s", res.StatusCode, string(b))
	}

	var data jiraSearchResult
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	set := map[string]struct{}{}
	for _, iss := range data.Issues {
		for _, l := range iss.Fields.Labels {
			if l == "" {
				continue
			}
			set[l] = struct{}{}
		}
	}

	labels := make([]string, 0, len(set))
	for k := range set {
		labels = append(labels, k)
	}
	sort.Strings(labels)
	return labels, nil
}
