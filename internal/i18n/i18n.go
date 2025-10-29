package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

var messages map[string]string
var currentLang string = "en"

func LoadLanguage(lang string) error {
	var data []byte
	switch lang {
	case "de":
		data = LocaleDE
	case "en":
		data = LocaleEN
	default:
		return fmt.Errorf("unsupported language: %s", lang)
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse embedded language data: %w", err)
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

//go:embed messages_de.json
var LocaleDE []byte

//go:embed messages_en.json
var LocaleEN []byte
