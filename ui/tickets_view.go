package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/helper"
	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
	"github.com/scramb/backlog-manager/ui/components"
)

// NewTicketsView builds the “My Tickets” tab content.
// It lists all issues assigned to the logged-in user and allows opening them in the browser.
func TicketsView(app fyne.App, w fyne.Window, domain, user, token string, reloadChan <-chan bool) fyne.CanvasObject {
	var issues []models.JiraIssue
	filteredIssues := []models.JiraIssue{}
	selectedProject := i18n.T("tickets.all_projects")

	ticketList := widget.NewList(
		func() int { return len(filteredIssues) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(nil)
			id := widget.NewLabel("ID")
			title := widget.NewLabel("Titel")
			openBtn := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), nil)
			box := container.NewHBox(icon, id, title, layout.NewSpacer(), openBtn)
			return box
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			issue := filteredIssues[i]
			box := o.(*fyne.Container)

			iconWidget := box.Objects[0].(*widget.Icon)
			idLabel := box.Objects[1].(*widget.Label)
			titleLabel := box.Objects[2].(*widget.Label)
			openBtn := box.Objects[4].(*widget.Button)

			var icon fyne.Resource
			switch issue.Fields.IssueType.Name {
			case "Epic":
				icon = theme.GridIcon()
			case "Story":
				icon = theme.FileIcon()
			case "Bug":
				icon = theme.ErrorIcon()
			case "Task":
				icon = theme.DocumentIcon()
			default:
				icon = theme.InfoIcon()
			}

			iconWidget.SetResource(icon)
			idLabel.SetText(issue.Key)
			titleLabel.SetText(issue.Fields.Summary)

			openBtn.OnTapped = func() {
				url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", domain, issue.Key)
				helper.OpenBrowser(url)
			}
		},
	)

	searchQuery := ""

	applyFilter := func(project string) {
		selectedProject = project
		filteredIssues = nil
		for _, iss := range issues {
			if project != i18n.T("tickets.all_projects") && !strings.HasPrefix(iss.Key, project+"-") {
				continue
			}
			if searchQuery != "" {
				lowerQuery := strings.ToLower(searchQuery)
				if !strings.Contains(strings.ToLower(iss.Key), lowerQuery) && !strings.Contains(strings.ToLower(iss.Fields.Summary), lowerQuery) {
					continue
				}
			}
			filteredIssues = append(filteredIssues, iss)
		}
		ticketList.Refresh()
	}

	projectFilter := widget.NewSelect([]string{i18n.T("tickets.all_projects")}, func(selected string) {
		applyFilter(selected)
	})
	projectFilter.SetSelected(i18n.T("tickets.all_projects"))

	// Move these widget creations above showListView()
	reloadBtn := i18n.BindButton("tickets.reload", theme.ViewRefreshIcon(), nil)
	projectFilterLabel := i18n.BindLabel("tickets.project_filter")
	searchEntryWidget := i18n.BindEntryWithPlaceholder("tickets.search_placeholder", false)

	searchEntryWidget.OnChanged = func(text string) {
		searchQuery = text
		applyFilter(selectedProject)
	}

	reloadBtn.OnTapped = func() {
		go func() {
			fetched, err := models.FetchAssignedIssues(domain, user, token)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() {
				issues = fetched
				projectSet := map[string]struct{}{}
				for _, iss := range issues {
					parts := strings.SplitN(iss.Key, "-", 2)
					if len(parts) > 1 {
						projectSet[parts[0]] = struct{}{}
					}
				}
				projects := []string{i18n.T("tickets.all_projects")}
				for p := range projectSet {
					projects = append(projects, p)
				}
				sort.Strings(projects)
				projectFilter.Options = projects
				projectFilter.Refresh()
				applyFilter(selectedProject)
			})
		}()
	}

	var contentContainer *fyne.Container

	showListView := func() {
		contentContainer.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					reloadBtn,
					projectFilterLabel,
					projectFilter,
					searchEntryWidget,
				),
				nil, nil, nil,
				ticketList,
			),
		}
		contentContainer.Refresh()
	}

	ticketList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(filteredIssues) {
			return
		}
		issue := filteredIssues[id]
		contentContainer.Objects = []fyne.CanvasObject{
			TicketDetailView(app, w, issue, domain, user, token, showListView),
		}
		contentContainer.Refresh()
	}

	contentContainer = container.NewMax()
	showListView()

	// Initial auto-refresh when tab is opened (optional)
	content := contentContainer.Objects[0].(*fyne.Container)
	if btn, ok := content.Objects[0].(*widget.Button); ok {
		btn.OnTapped()
	}

	// Add language change handler to update the "Alle Projekte" text dynamically
	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			projectFilterLabel.SetText(i18n.T("tickets.project_filter"))
			reloadBtn.SetText(i18n.T("tickets.reload"))
			searchEntryWidget.SetPlaceHolder(i18n.T("tickets.search_placeholder"))

			prevSelection := projectFilter.Selected
			translatedAll := i18n.T("tickets.all_projects")

			newOptions := []string{translatedAll}
			for _, opt := range projectFilter.Options {
				if opt != "All Projects" && opt != "Alle Projekte" {
					newOptions = append(newOptions, opt)
				}
			}

			newSelect := widget.NewSelect(newOptions, func(selected string) {
				applyFilter(selected)
			})

			if prevSelection == "All Projects" || prevSelection == "Alle Projekte" {
				newSelect.SetSelected(translatedAll)
			} else {
				newSelect.SetSelected(prevSelection)
			}

			// Versuche, den Container sicher zu ersetzen
			if len(contentContainer.Objects) == 0 {
				return
			}

			if content, ok := contentContainer.Objects[0].(*fyne.Container); ok {
				if len(content.Objects) > 0 {
					if border, ok := content.Objects[0].(*fyne.Container); ok {
						if len(border.Objects) > 0 {
							if header, ok := border.Objects[0].(*fyne.Container); ok {
								for i, obj := range header.Objects {
									if obj == projectFilter {
										header.Objects[i] = newSelect
										projectFilter = newSelect
										contentContainer.Refresh()
										return
									}
								}
							}
						}
					}
				}
			}

			// 🧹 Falls der View gerade nicht der erwartete Container ist
			// → einfach komplette Ansicht neu aufbauen
			showListView()
		})
	})

	go func() {
		for range reloadChan {
			fyne.Do(func() {
				fetched, err := models.FetchAssignedIssues(domain, user, token)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				issues = fetched
				// build project list
				projectSet := map[string]struct{}{}
				for _, iss := range issues {
					parts := strings.SplitN(iss.Key, "-", 2)
					if len(parts) > 1 {
						projectSet[parts[0]] = struct{}{}
					}
				}
				projects := []string{i18n.T("tickets.all_projects")}
				for p := range projectSet {
					projects = append(projects, p)
				}
				sort.Strings(projects)
				projectFilter.Options = projects
				projectFilter.Refresh()

				applyFilter(selectedProject)
			})
		}
	}()

	return contentContainer
}





// TicketDetailView shows detailed information about a Jira issue with a back button.
func TicketDetailView(app fyne.App, w fyne.Window, issue models.JiraIssue, domain, user, token string, back func()) fyne.CanvasObject {
	keyLabel := widget.NewLabelWithStyle(issue.Key, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labels, transitions, comments := loadTicketContent(issue, domain, user, token)
	summaryHeader := i18n.BindLabel("tickets.summary")
	summaryLabel := widget.NewLabel(issue.Fields.Summary)

	descriptionHeader := i18n.BindLabel("tickets.description")
	description := models.ExtractDescriptionText(issue.Fields.Description)
	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	// Labels-Bereich vorbereiten
	labelTitle := widget.NewLabel("Labels:")
	labelsFlow := container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 30)))

	go func() {
		fyne.Do(func() {
			for _, lbl := range labels.Fields.Labels {
				labelsFlow.Add(components.CreateChip(lbl))
			}
			labelsFlow.Refresh()
		})
	}()

	backBtn := i18n.BindButton("tickets.back", theme.NavigateBackIcon(), func() {
		back()
	})
	transitionOptions := []string{}
	transitionMap := map[string]string{}
	for _, t := range transitions {
		transitionOptions = append(transitionOptions, t.Name)
		transitionMap[t.Name] = t.ID
	}

	var selectedTransitionID string

	transitionSelect := widget.NewSelect(transitionOptions, func(selected string) {
		selectedTransitionID = transitionMap[selected]
		fmt.Println(selectedTransitionID)
	})

	transitionSelect.Disable()

	transitionContainer := container.NewVBox(i18n.BindLabel("tickets.transition_label"), transitionSelect)

	transitionSelect.OnChanged = func(selected string) {
		fmt.Printf("Selected: %s", transitionMap[selected])
	}

	commentsContainer := container.NewVBox()
	addCommentSection := widget.NewMultiLineEntry()
	addCommentBtn := i18n.BindButton("tickets.add_comment_button", nil, nil)

	commentsContainer.Add(i18n.BindLabel("tickets.add_comment_header"))
	commentsContainer.Add(addCommentSection)
	commentsContainer.Add(addCommentBtn)

	for _, c := range comments {
		commentsContainer.Add(components.CreateChatMessageCard(c, user))
	}

	addCommentBtn.OnTapped = func() {
		addCommentBtn.Disable()
		go func () {
			if addCommentSection.Text == "" {
				fyne.Do(func() {
					addCommentBtn.Enable()
					dialog.ShowInformation(i18n.T("tickets.error"), i18n.T("tickets.error_fields"), w)
				})
				return
			}
			err := models.AddCommentToTicket(domain, user, token, issue.Id, addCommentSection.Text)
			fyne.Do(func() {
				addCommentBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				dialog.ShowInformation(i18n.T("tickets.comment_created"), i18n.T("tickets.created"), w)
				addCommentSection.SetText("")
			})
		}()
	}

	detailsSection := components.CollapsibleSection(i18n.T("tickets.comment_section_header"), commentsContainer)

	content := container.NewVBox(
		backBtn,
		keyLabel,
		widget.NewSeparator(),
		transitionContainer,
		summaryHeader,
		summaryLabel,
		widget.NewSeparator(),
		descriptionHeader,
		descriptionLabel,
		labelTitle,
		labelsFlow, // statt labelsContainer
		detailsSection,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	return scroll
}

func loadTicketContent(issue models.JiraIssue, domain, user, token string) (models.JiraIssueLabels, []models.JiraTransition, []models.JiraComment) {

	labels, errLabels := models.FetchIssueLabels(domain, user, token, issue.Id)
	comments, errComments := models.FetchIssueComments(domain, user, token, issue.Id)
	transitions, errTransitions := models.FetchIssueTransitions(domain, user, token, issue.Id)

	if errComments != nil || errLabels != nil || errTransitions != nil {
		fmt.Print(errLabels, errComments, errTransitions)
	}

	return labels, transitions, comments
}

