<p align="center">
  <img src="https://github.com/scramb/Jirion/blob/main/assets/logo_cropped.png?raw=true" alt="Jirion Logo"/>
</p>

# Jirion

A modern, cross-platform **Jira client** for macOS, Windows, and Linux – built with **Go** and **Fyne**.  
Jirion helps you **manage tickets**, **create backlog items**, and **organize projects clearly** – without the typical Jira overhead.

---

## 🚀 Features

- 🧙 **Setup Wizard** – guided initial setup for Jira domain, API token & user.
- 🧱 **Create Backlog Items** – create new tickets directly, including type, title, description **and labels**.
- 🏷️ **Label Management** – load Jira labels per project, select your favorites & save them permanently.
- 🔄 **My Tickets View** – see all issues assigned to you at a glance.
- 🤖 **AI Suggestions (optional)** – use OpenAI-compatible APIs for description generation.
- 💾 **Persistent Configuration** – all data is saved automatically (Fyne preferences system).
- 💡 **Cross-Platform Builds** – runs natively on macOS, Windows & Linux (AMD64 + ARM64).

---

## 🧩 Project Structure

```
backlog-manager/
├── main.go                      # Entry point, Setup Wizard & app initialization
├── ui/
│   ├── settings                 # SubViews for Settings
│   ├── backlog_view.go          # Create Backlog View (incl. label selection)
│   ├── tickets_view.go          # My Tickets View + detail view
│   ├── settings_view.go         # Settings & label config (persisted per project)
│   ├── setup_wizard.go          # Setup Wizard for Jira config
│   └── ...
├── internal/models/             # Jira API logic (requests, CreateIssue, etc.)
├── internal/i18n/               # i18n logic
├── internal/components/         # Components 
├── assets/                      # App icons & static resources
├── go.mod                       # Go module definition
└── go.sum
```

---

## ⚙️ Installation & Development

### Requirements
- [Go 1.21+](https://go.dev/dl/)
- Git
- [Fyne Toolkit](https://developer.fyne.io/)

### Local Setup
```bash
# Clone repository
git clone https://github.com/scramb/jirion.git
cd jirion

# Install dependencies
go mod tidy

# Start app (dev mode)
go run .

# Production build (macOS example)
fyne package -release -os darwin -icon ./assets/app.png -name Jirion -app-id com.scramb.jirion
```

### Development Mode (persistent data)
By default, Fyne stores preferences on macOS here:
```
~/Library/Preferences/fyne/jirion/preferences.json
```
For a repo-local dev store, you can (optionally) set in `main.go`:
```go
os.Setenv("FYNE_APP_STORAGE", "./.fyne")
```
Then data will be stored at:
```
./.fyne/preferences.json
```

---

## 📦 Release Builds (via GitHub Actions)

On release, the following are automatically created:
- `jirion-darwin-amd64.dmg`
- `jirion-darwin-arm64.dmg`
- `jirion-windows-amd64.exe`
- `jirion-linux-amd64`

---

## 🧠 Usage

### First Start
On first launch, the **Setup Wizard** opens automatically.  
Enter your Jira instance (e.g. `<jira-space>.atlassian.net`), your email, and your API token.

### Main View
- **Create Backlog** → Create new ticket (type, title, description, **labels**).
- **My Tickets** → View your current tasks (+ ticket detail page).
- **Settings** → AI endpoint, system prompt & **label configuration per project**.

---

## 📸 Screenshots 

| Setup Wizard | Tickets View | Backlog Creation |
|--------------|--------------|------------------|
| ![Setup](https://i.ibb.co/FkmzzM8G/Bildschirmfoto-2025-10-24-um-23-32-05.png) | ![Tickets](https://i.ibb.co/Yn9GD2t/Bildschirmfoto-2025-10-24-um-23-21-35.png) | ![Create](https://i.ibb.co/q3sZ5S2H/Bildschirmfoto-2025-10-24-um-23-21-29.png) |
| Settings View | Tickets Detail View | ServiceDesk View (experimental) |
|--------------|--------------|------------------|
| ![Setup](https://i.ibb.co/spcTn9xp/Bildschirmfoto-2025-10-31-um-08-09-55.png) | ![Tickets](https://i.ibb.co/WpDQzy9y/Bildschirmfoto-2025-10-31-um-08-12-00.png) | ![Create](https://i.ibb.co/tTT6MHVM/Bildschirmfoto-2025-10-31-um-08-09-17.png) |
---

## 💬 Contact

**Author:** Carsten Meininger  
**GitHub:** [@scramb](https://github.com/scramb)  
**E-Mail:** carschi92@gmail.com

---

## License – MIT License + Restricted Commercial Use Addendum

MIT License

Copyright (c) 2025 Carsten Meininger

Permission is hereby granted, free of charge, to any person obtaining a copy  
of this software and associated documentation files (the "Software"), to deal  
in the Software without restriction, including without limitation the rights  
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell  
copies of the Software, and to permit persons to whom the Software is  
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all  
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR  
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,  
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE  
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER  
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,  
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE  
SOFTWARE.

---

### Restricted Commercial Use Addendum

Use of this software for commercial purposes by entities with more than 50 employees or annual revenues exceeding $1 million USD requires a separate commercial license from the author. For licensing inquiries, please contact Carsten Meininger at carschi92@gmail.com.

---

> Jirion – because Jira deserves a better UX._
