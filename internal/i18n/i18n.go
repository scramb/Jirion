package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var messages map[string]string
var currentLang string = "en"

func LoadLanguage(lang string) error {
	configDir, _ := os.Getwd()
	path := filepath.Join(configDir, "internal", "i18n", fmt.Sprintf("messages_%s.json", lang))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not load language file: %w", err)
	}
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse language file: %w", err)
	}
	currentLang = lang
	notifyLanguageChange()
	return nil
}

func T(key string) string {
	if val, ok := messages[key]; ok {
		return val
	}
	return fmt.Sprintf("{{%s}}", key)
}

func CurrentLanguage() string {
	return currentLang
}

var onLanguageChangeCallbacks []func()

func RegisterOnLanguageChange(cb func()) {
	onLanguageChangeCallbacks = append(onLanguageChangeCallbacks, cb)
}

func notifyLanguageChange() {
	for _, cb := range onLanguageChangeCallbacks {
		cb()
	}
}
