package services

// ServiceInfo contains information about a detected service
type ServiceInfo struct {
	Name     string
	Icon     string // Simple Icons CDN URL for SVG icon
	Type     string // "docs", "sheets", "notes", "wiki", etc.
	Referrer string // Original referrer parameter value
}

// DetectServiceFromReferrer detects the service from a 'referrer' parameter
func DetectServiceFromReferrer(referrerParam string) *ServiceInfo {
	if referrerParam == "" {
		return nil
	}

	switch referrerParam {
	case "google-docs":
		return &ServiceInfo{Name: "Google Docs", Icon: "https://cdn.simpleicons.org/googledocs", Type: "docs", Referrer: referrerParam}
	case "google-sheets":
		return &ServiceInfo{Name: "Google Sheets", Icon: "https://cdn.simpleicons.org/googlesheets", Type: "sheets", Referrer: referrerParam}
	case "google-slides":
		return &ServiceInfo{Name: "Google Slides", Icon: "https://cdn.simpleicons.org/googleslides", Type: "presentation", Referrer: referrerParam}
	case "google-drive":
		return &ServiceInfo{Name: "Google Drive", Icon: "https://cdn.simpleicons.org/googledrive", Type: "storage", Referrer: referrerParam}
	case "google":
		return &ServiceInfo{Name: "Google", Icon: "https://cdn.simpleicons.org/google", Type: "google", Referrer: referrerParam}
	case "notion":
		return &ServiceInfo{Name: "Notion", Icon: "https://cdn.simpleicons.org/notion", Type: "notes", Referrer: referrerParam}
	case "confluence":
		return &ServiceInfo{Name: "Confluence", Icon: "https://cdn.simpleicons.org/confluence", Type: "wiki", Referrer: referrerParam}
	case "microsoft":
		return &ServiceInfo{Name: "Microsoft Office", Icon: "https://cdn.simpleicons.org/microsoft", Type: "office", Referrer: referrerParam}
	case "github":
		return &ServiceInfo{Name: "GitHub", Icon: "https://cdn.simpleicons.org/github", Type: "code", Referrer: referrerParam}
	case "gitlab":
		return &ServiceInfo{Name: "GitLab", Icon: "https://cdn.simpleicons.org/gitlab", Type: "code", Referrer: referrerParam}
	case "outline":
		return &ServiceInfo{Name: "Outline", Icon: "https://cdn.simpleicons.org/outline", Type: "wiki", Referrer: referrerParam}
	case "slack":
		return &ServiceInfo{Name: "Slack", Icon: "https://cdn.simpleicons.org/slack", Type: "chat", Referrer: referrerParam}
	case "discord":
		return &ServiceInfo{Name: "Discord", Icon: "https://cdn.simpleicons.org/discord", Type: "chat", Referrer: referrerParam}
	case "trello":
		return &ServiceInfo{Name: "Trello", Icon: "https://cdn.simpleicons.org/trello", Type: "boards", Referrer: referrerParam}
	case "asana":
		return &ServiceInfo{Name: "Asana", Icon: "https://cdn.simpleicons.org/asana", Type: "tasks", Referrer: referrerParam}
	case "monday":
		return &ServiceInfo{Name: "Monday.com", Icon: "https://cdn.simpleicons.org/monday", Type: "project", Referrer: referrerParam}
	case "figma":
		return &ServiceInfo{Name: "Figma", Icon: "https://cdn.simpleicons.org/figma", Type: "design", Referrer: referrerParam}
	case "miro":
		return &ServiceInfo{Name: "Miro", Icon: "https://cdn.simpleicons.org/miro", Type: "whiteboard", Referrer: referrerParam}
	case "dropbox":
		return &ServiceInfo{Name: "Dropbox", Icon: "https://cdn.simpleicons.org/dropbox", Type: "storage", Referrer: referrerParam}

	default:
		return &ServiceInfo{Name: referrerParam, Icon: "https://cdn.simpleicons.org/link", Type: "custom", Referrer: referrerParam}
	}
}
