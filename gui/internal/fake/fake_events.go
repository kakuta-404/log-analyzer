package fake

import (
	"GUI/internal/models"
	"strconv"
	"time"
)

func GenerateFakeEvents(projectID string, count int) []models.Event {
	var events []models.Event

	names := []string{
		"login", "signup", "click", "download", "upload", "logout",
		"search", "error", "view", "navigate", "name10", "name11",
	}
	roles := []string{"admin", "user", "guest"}
	devices := []string{"mobile", "desktop", "tablet"}
	browsers := []string{"Firefox", "Chrome", "Safari", "Edge"}

	for i := 0; i < count; i++ {
		name := names[i%len(names)]
		role := roles[i%len(roles)]
		device := devices[i%len(devices)]
		browser := browsers[i%len(browsers)]
		userID := strconv.Itoa((i % 7) + 1000) // e.g., 1000–1006

		events = append(events, models.Event{
			ID:         "e" + strconv.Itoa(i),
			Name:       name,
			Timestamp:  time.Now().Add(-time.Duration(i) * time.Minute).Format("2006-01-02 15:04"),
			InsertedAt: time.Now().Add(-time.Duration(i) * time.Minute).Format("2006-01-02 15:05"),

			SearchableKeys: map[string]string{
				"user_id": userID,
				"role":    role,
				"device":  device,
			},

			OtherKeys: map[string]string{
				"ip":      "192.168.1." + strconv.Itoa(i%255),
				"browser": browser,
			},
		})
	}
	return events
}
