package generator

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/kakuta-404/log-analyzer/common"
	"github.com/kakuta-404/log-analyzer/log-generator/internal/projects"
)

type Generator struct {
	logDrainURL string
	projectSvc  *projects.Service
	client      *http.Client
}

func New(cfg struct{ LogDrainURL string }, projectSvc *projects.Service) *Generator {
	return &Generator{
		logDrainURL: cfg.LogDrainURL,
		projectSvc:  projectSvc,
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (g *Generator) GenerateAndSend() error {
	projects, err := g.projectSvc.GetProjects()
	if err != nil || len(projects) == 0 {
		return err
	}
	p := projects[rand.Intn(len(projects))]

	sub := common.Submission{
		ProjectID: p.ID,
		APIKey:    p.APIKey,
		Name:      randomString(),
		Timestamp: time.Now(),
		PayLoad:   randomPayload(p.SearchableKeys),
	}

	body, _ := json.Marshal(sub)
	resp, err := g.client.Post(g.logDrainURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func randomString() string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	return string(letters[rand.Intn(len(letters))])
}

func randomPayload(keys []string) map[string]string {
	payload := make(map[string]string)
	for _, k := range keys {
		payload[k] = randomString()
	}
	payload["registered"] = map[bool]string{true: "true", false: "false"}[rand.Intn(2) == 1]
	payload["extra"] = randomString()
	return payload
}
