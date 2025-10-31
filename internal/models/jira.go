package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type JiraIssue struct {
	Key    string `json:"key"`
	Id     string `json:"id"`
	Fields struct {
		Summary     string          `json:"summary"`
		Description json.RawMessage `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

type JiraIssueDetails struct {
	Labels      JiraIssueLabels  `json:"fields"`
	Comments    []JiraComment    `json:"comments"`
	Transitions []JiraTransition `json:"transitions"`
}

type JiraIssueLabels struct {
	Fields struct {
		Labels []string `json:"labels"`
	}
}

type JiraCommentResult struct {
	Comments []JiraComment `json:comment`
}

type JiraComment struct {
	Author struct {
		Email       string `json:"emailAddress"`
		DisplayName string `json:"displayName"`
		AvatarUrls  struct {
			Image string `json:"48x48"`
		}
	}
	Content json.RawMessage `json:"body"`
}

type JiraTransitionResult struct {
	Transitions []JiraTransition `json:"transitions"`
}

type JiraTransition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

func FetchIssueTransitions(domain, email, token string, id string) ([]JiraTransition, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue/%s/transitions", domain, id)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
	}

	var result JiraTransitionResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Transitions, nil
}

func FetchAssignedIssues(domain, email, apiToken string) ([]JiraIssue, error) {
	jql := `assignee=currentUser() AND status NOT IN ("Done", "Canceled", "Cancelled", "Approved")`
	encodedJQL := url.QueryEscape(jql)
	params := "fields=id,summary,issuetype,key,description"
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/search/jql?jql=%s&%s", domain, encodedJQL, params)
	req, _ := http.NewRequest("GET", url, nil)

	decryptedToken := TryDecrypt(apiToken)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
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

func FetchIssueComments(domain, email, token string, id string) ([]JiraComment, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue/%s/comment", domain, id)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
	}

	var result JiraCommentResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Comments, nil
}

func FetchIssueLabels(domain, email, token string, id string) (JiraIssueLabels, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/issue/%s?fields=labels", domain, id)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return JiraIssueLabels{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return JiraIssueLabels{}, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
	}

	var result JiraIssueLabels
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return JiraIssueLabels{}, err
	}
	return result, nil
}

func FetchFavouriteProjects(domain, email, token string) ([]JiraProject, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/project/search?favourite=true", domain)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
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
	decryptedToken := TryDecrypt(token)
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
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(body))
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
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(b))
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
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira api error (%d): %s", res.StatusCode, string(b))
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

// JiraServiceDesk represents a service desk within Jira Service Management
type JiraServiceDesk struct {
	ID   string `json:"id"`
	Name string `json:"projectName"`
}

type JiraServiceDeskList struct {
	Values []JiraServiceDesk `json:"values"`
}

// FetchServiceDesks returns all visible service desks for the authenticated user
func FetchServiceDesks(domain, email, token string) ([]JiraServiceDesk, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/servicedeskapi/servicedesk", domain)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira service desk api error (%d): %s", res.StatusCode, string(b))
	}

	var out JiraServiceDeskList
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Values, nil
}

// JiraServiceRequest represents a created or retrieved service request
type JiraServiceRequest struct {
	IssueKey string `json:"issueKey"`
	IssueID  string `json:"issueId"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
}

// CreateServiceRequest creates a new request in a service desk
func CreateServiceRequest(domain, email, token, serviceDeskID, requestTypeID, summary, description string, fields map[string]interface{}) (*JiraServiceRequest, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/servicedeskapi/request", domain)

	payload := map[string]interface{}{
		"serviceDeskId": serviceDeskID,
		"requestTypeId": requestTypeID,
		"requestFieldValues": map[string]interface{}{
			"summary":     summary,
			"description": description,
		},
	}

	for k, v := range fields {
		payload["requestFieldValues"].(map[string]interface{})[k] = v
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	decryptedToken := TryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to create service request: %s", string(b))
	}

	var out JiraServiceRequest
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

// FetchMyServiceRequests returns all open service requests created by the current user
func FetchMyServiceRequests(domain, email, token string) ([]JiraServiceRequest, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/servicedeskapi/request?requestOwnership=OWNED_REQUESTS&requestStatus=ALL_REQUESTS", domain)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira service desk api error (%d): %s", res.StatusCode, string(b))
	}

	var data struct {
		Values []struct {
			IssueKey      string `json:"issueKey"`
			IssueID       string `json:"issueId"`
			Summary       string `json:"summary"`
			CurrentStatus struct {
				Name string `json:"name"`
			} `json:"currentStatus"`
		} `json:"values"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	var out []JiraServiceRequest
	for _, v := range data.Values {
		out = append(out, JiraServiceRequest{
			IssueKey: v.IssueKey,
			IssueID:  v.IssueID,
			Summary:  v.Summary,
			Status:   v.CurrentStatus.Name,
		})
	}

	return out, nil
}

// FetchRequestComments retrieves comments for a specific service request
func FetchRequestComments(domain, email, token, issueID string) ([]string, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/servicedeskapi/request/%s/comment", domain, issueID)
	req, _ := http.NewRequest("GET", url, nil)
	decryptedToken := TryDecrypt(token)
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
		return nil, fmt.Errorf("jira service desk api error (%d): %s", res.StatusCode, string(b))
	}

	var data struct {
		Values []struct {
			Body struct {
				Content []struct {
					Content []struct {
						Text string `json:"text"`
					} `json:"content"`
				} `json:"content"`
			} `json:"body"`
		} `json:"values"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	var comments []string
	for _, c := range data.Values {
		for _, cc := range c.Body.Content {
			for _, text := range cc.Content {
				comments = append(comments, text.Text)
			}
		}
	}

	return comments, nil
}

// AddCommentToRequest adds a new comment to a specific service request
func AddCommentToRequest(domain, email, token, issueID, message string) error {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/servicedeskapi/request/%s/comment", domain, issueID)
	payload := map[string]interface{}{
		"body": message,
	}
	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	decryptedToken := TryDecrypt(token)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, decryptedToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to add comment: %s", string(b))
	}

	return nil
}
