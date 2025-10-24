

# 🍺 Backlog Manager

Ein moderner, plattformübergreifender **Jira-Client** für macOS, Windows und Linux – entwickelt mit **Go** und **Fyne**.  
Der Backlog Manager hilft dir, **Tickets zu verwalten**, **Backlog-Items zu erstellen** und **Projekte übersichtlich zu organisieren** – ohne den typischen Jira-Overhead.

---

## 🚀 Features

- 🧙 **Setup Wizard** – geführte Ersteinrichtung für Jira-Domain, API-Token & Benutzer.
- 🧱 **Create Backlog Items** – neue Tickets direkt anlegen, inkl. Typ, Titel, Beschreibung **und Labels**.
- 🏷️ **Label Management** – lade Jira-Labels pro Projekt, wähle deine Favoriten & speichere sie dauerhaft.
- 🔄 **My Tickets View** – zeig auf einen Blick alle dir zugewiesenen Issues.
- 🤖 **KI-Vorschläge (optional)** – nutze OpenAI-kompatible APIs zur Beschreibungserstellung.
- 💾 **Persistente Konfiguration** – alle Daten werden automatisch gespeichert (Preferences-System von Fyne).
- 💡 **Cross-Platform Builds** – läuft nativ auf macOS, Windows & Linux (AMD64 + ARM64).

---

## 🧩 Projektstruktur

```
backlog-manager/
├── main.go                      # Einstiegspunkt, Setup Wizard & App Initialisierung
├── ui/
│   ├── backlog_view.go          # Create Backlog View (inkl. Label-Auswahl)
│   ├── tickets_view.go          # My Tickets View + Detailansicht
│   ├── settings_view.go         # Settings & Label Config (pro Projekt persistiert)
│   ├── setup_wizard.go          # Setup Wizard für Jira-Config
│   └── ...
├── internal/models/             # Jira API Logik (Requests, CreateIssue, etc.)
├── assets/                      # App-Icons & statische Ressourcen
├── go.mod                       # Go Module Definition
└── go.sum
```

---

## ⚙️ Installation & Entwicklung

### Voraussetzungen
- [Go 1.21+](https://go.dev/dl/)
- Git
- [Fyne Toolkit](https://developer.fyne.io/)

### Lokales Setup
```bash
# Repository klonen
git clone https://github.com/scramb/backlog-manager.git
cd backlog-manager

# Abhängigkeiten installieren
go mod tidy

# App starten (Dev Mode)
go run .

# Produktionsbuild (macOS Beispiel)
go build -o backlog-manager .
```

### Entwicklungsmodus (persistente Daten)
Standardmäßig speichert Fyne die Preferences unter macOS hier:
```
~/Library/Preferences/fyne/backlog-manager/preferences.json
```
Für einen repo-lokalen Dev-Store kannst du (optional) in `main.go` setzen:
```go
os.Setenv("FYNE_APP_STORAGE", "./.fyne")
```
Dann liegen die Daten unter:
```
./.fyne/preferences.json
```

---

## 📦 Release-Builds (via GitHub Actions)

Bei einem Release werden automatisch erstellt:
- `backlog-manager-darwin-amd64.dmg`
- `backlog-manager-darwin-arm64.dmg`
- `backlog-manager-windows-amd64.exe`
- `backlog-manager-linux-amd64`

---

## 🧠 Nutzung

### Erster Start
Beim ersten Start öffnet sich automatisch der **Setup Wizard**.  
Trage dort deine Jira-Instanz (z. B. `<jira-space>.atlassian.net`), deine E-Mail und dein API-Token ein.

### Hauptansicht
- **Create Backlog** → Neues Ticket anlegen (Typ, Titel, Beschreibung, **Labels**).
- **My Tickets** → Deine aktuellen Aufgaben anzeigen (+ Ticket-Detailseite).
- **Settings** → KI-Endpoint, System-Prompt & **Label-Konfiguration pro Projekt**.

---

## 📸 Screenshots *(optional)*

| Setup Wizard | Tickets View | Backlog Creation |
|--------------|--------------|------------------|
| ![Setup](assets/screens/setup.png) | ![Tickets](assets/screens/tickets.png) | ![Create](assets/screens/backlog.png) |

---

## 💬 Kontakt

**Autor:** Carsten Meininger  
**GitHub:** [@scramb](https://github.com/scramb)  
**E-Mail:** carschi92@gmail.com

---

## 🍺 License – Beerware License (Revision 42)

```
"THE BEERWARE LICENSE" (Revision 42):
<carschi92@gmail.com> wrote this software. As long as you retain this notice,
you can do whatever you want with this stuff. If we meet someday, and you think this
stuff is worth it, you can buy me a beer in return.
```

> _Backlog Manager – because Jira deserves a better UX._