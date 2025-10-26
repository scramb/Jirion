package settings

import (
	"fmt"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BuildAISettings builds the AI configuration settings tab.
func BuildAISettings(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	enableAI := widget.NewCheck(i18n.T("settings.enable_ai_features"), nil)
	enableAI.SetChecked(prefs.Bool("ai_enabled"))

	endpointEntry := i18n.BindEntryWithPlaceholder("settings.ai_endpoint_placeholder", false)
	endpointEntry.SetText(prefs.String("ai_endpoint"))
	systemPromptEntry := widget.NewMultiLineEntry()
	systemPromptEntry.SetPlaceHolder(i18n.T("settings.system_prompt_placeholder"))
	systemPromptEntry.Wrapping = fyne.TextWrapWord
	systemPromptEntry.SetText(prefs.String("system_prompt"))
	systemPromptEntry.ExtendBaseWidget(systemPromptEntry)
	apiKeyEntry := i18n.BindEntryWithPlaceholder("settings.api_key_placeholder", true)

    // Prefill API key once, if stored and entry is empty
    if enc := prefs.String("openai_api_key"); enc != "" {
        if dec := models.TryDecrypt(enc); dec != "" && apiKeyEntry.Text == "" {
            apiKeyEntry.SetText(dec)
        }
    }

	// Disable inputs when AI features are turned off
	updateFields := func(enabled bool) {
		endpointEntry.Disable()
		systemPromptEntry.Disable()
		apiKeyEntry.Disable()
		if enabled {
			endpointEntry.Enable()
			systemPromptEntry.Enable()
			apiKeyEntry.Enable()
            // Prefill on enable as a fallback (no focus callback)
            if enc := prefs.String("openai_api_key"); enc != "" {
                if dec := models.TryDecrypt(enc); dec != "" && apiKeyEntry.Text == "" {
                    apiKeyEntry.SetText(dec)
                }
            }
		}
	}

	enableAI.OnChanged = func(checked bool) {
		prefs.SetBool("ai_enabled", checked)
		updateFields(checked)
	}

	updateFields(prefs.Bool("ai_enabled"))

	saveBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		if !enableAI.Checked {
			prefs.SetBool("ai_enabled", false)
			dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.ai_disabled"), w)
			return
		}

		prefs.SetBool("ai_enabled", true)
		prefs.SetString("ai_endpoint", endpointEntry.Text)
		prefs.SetString("system_prompt", systemPromptEntry.Text)

		if apiKeyEntry.Text != "" {
			encryptedKey, err := models.Encrypt(apiKeyEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.T("settings.error_encrypt_api_key")+": %w", err), w)
				return
			}
			prefs.SetString("openai_api_key", encryptedKey)
		}

		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.ai_saved"), w)
	})

	formContent := container.NewVBox(
		enableAI,
		i18n.BindLabel("settings.ai_endpoint"),
		endpointEntry,
		i18n.BindLabel("settings.system_prompt"),
		systemPromptEntry,
		i18n.BindLabel("settings.api_key"),
		apiKeyEntry,
	)

	scroll := container.NewVScroll(formContent)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	pinnedSave := container.NewBorder(nil, saveBtn, nil, nil, scroll)

	return pinnedSave
}