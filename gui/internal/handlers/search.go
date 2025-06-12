package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"net/http"
)

func ShowSearchPage(c *gin.Context) {
	user, ok := getAuthenticatedUser(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	projectName, searchKeys, ok := validateProjectAccess(c, user, projectID)
	if !ok {
		return
	}

	filters := map[string]string{}
	if c.Request.Method == http.MethodPost {
		for _, key := range searchKeys {
			val := c.PostForm(key)
			if val != "" {
				filters[key] = val
			}
		}
	}

	// Build REST API request
	url := fmt.Sprintf("%s/projects/%s/search", common.RESTAPIBaseURL, projectID)

	// Attach filters as query parameters
	q := url + "?"
	for k, v := range filters {
		q += fmt.Sprintf("%s=%s&", k, v)
	}

	resp, err := http.Get(q)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "Failed to fetch filtered events")
		return
	}
	defer resp.Body.Close()

	var result common.GroupedEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.String(http.StatusInternalServerError, "Invalid response from server")
		return
	}

	c.HTML(http.StatusOK, "search.gohtml", gin.H{
		"ProjectID":   projectID,
		"ProjectName": projectName,
		"Filters":     filters,
		"SearchKeys":  searchKeys,
		"Groups":      result.Groups,
	})
}
