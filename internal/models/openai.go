package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
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

type OpenAIModelList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// FetchAvailableModels fetches model IDs from the OpenAI-like /models endpoint.
func FetchAvailableModels(endpoint, apiKey string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/models", endpoint), nil)
	if err != nil {
		return nil, err
	}

	decryptedToken := strings.TrimSpace(tryDecrypt(apiKey))
	// Falls jemand versehentlich einen bereits verschlüsselten Wert gespeichert hat:
	if !strings.HasPrefix(decryptedToken, "sk-") {
		second := strings.TrimSpace(tryDecrypt(decryptedToken))
		if strings.HasPrefix(second, "sk-") {
			decryptedToken = second
		}
	}

	// Standard: OpenAI-kompatibel → Bearer + sk-...
	req.Header.Set("Authorization", "Bearer "+decryptedToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("model fetch failed: %s", string(body))
	}

	var modelList OpenAIModelList
	if err := json.NewDecoder(resp.Body).Decode(&modelList); err != nil {
		return nil, err
	}

	models := make([]string, len(modelList.Data))
	for i, m := range modelList.Data {
		models[i] = m.ID
	}
	return models, nil
}

func GenerateBacklogContent(apiKey, endpoint, systemPrompt, userPrompt string) (string, error) {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1" // fallback
	}
	if !strings.HasSuffix(endpoint, "/v1") && !strings.HasSuffix(endpoint, "/v1/") {
		endpoint = strings.TrimRight(endpoint, "/") + "/v1"
	}
	url := endpoint + "/chat/completions"

	prefs := fyne.CurrentApp().Preferences()
	selectedModel := prefs.String("openai_model")
	if selectedModel == "" {
		selectedModel = "azure-gpt-4.1"
	}

	payload := map[string]interface{}{
		"model": selectedModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	decryptedKey := tryDecrypt(apiKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", decryptedKey))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API Fehler (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var data OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Choices) == 0 {
		return "", fmt.Errorf("keine Antwort von AI")
	}

	return data.Choices[0].Message.Content, nil
}
