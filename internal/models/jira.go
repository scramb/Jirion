package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
)

type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary   string `json:"summary"`
		IssueType struct {
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

func FetchAssignedIssues(domain, email, apiToken string) ([]JiraIssue, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/search/jql?jql=assignee=currentUser()&fields=summary,issuetype,key", domain)
	req, _ := http.NewRequest("GET", url, nil)

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, apiToken)))
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

func FetchFavouriteProjects(domain, email, token string) ([]JiraProject, error) {
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/project/search?favourite=true", domain)
	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, token)))
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

func CreateJiraIssue(domain, email, token, projectKey, issueType, title, content string) error {
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
            "description": adf, // ✅ korrektes ADF-Objekt
        },
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        return err
    }

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, token)))
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