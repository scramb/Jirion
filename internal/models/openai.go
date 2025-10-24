package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GenerateBacklogContent(apiKey, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]interface{}{
		"model": "azure-gpt-4.1",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://litellm-test.az.de.bauhaus.intra/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("OpenAI API Fehler (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var data OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Choices) == 0 {
		return "", fmt.Errorf("keine Antwort von OpenAI")
	}

	return data.Choices[0].Message.Content, nil
}
