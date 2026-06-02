package pyxis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type imageData struct {
	ParsedData struct {
		Labels []label `json:"labels"`
	} `json:"parsed_data"`
}

type apiResponse struct {
	Data []imageData `json:"data"`
}

func GetGoVersion(imageRef string) (string, error) {
	registry, repo, err := parseImageRef(imageRef)
	if err != nil {
		return "", err
	}

	encodedRepo := url.PathEscape(repo)
	apiURL := fmt.Sprintf(
		"https://catalog.redhat.com/api/containers/v1/repositories/registry/%s/repository/%s/images?filter=repositories.tags.name==latest&include=data.parsed_data.labels&page_size=1",
		registry, encodedRepo,
	)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("pyxis API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pyxis API returned %d", resp.StatusCode)
	}

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no images found for %s", imageRef)
	}

	for _, l := range result.Data[0].ParsedData.Labels {
		if l.Name == "version" {
			return l.Value, nil
		}
	}

	return "", fmt.Errorf("no version label found for %s", imageRef)
}

func parseImageRef(ref string) (registry, repo string, err error) {
	ref = strings.TrimPrefix(ref, "docker://")

	// registry.redhat.io/ubi9/go-toolset:tag -> registry=registry.access.redhat.com repo=ubi9/go-toolset
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid image reference: %s", ref)
	}

	repo = parts[1]
	if idx := strings.LastIndex(repo, ":"); idx != -1 {
		repo = repo[:idx]
	}

	registry = "registry.access.redhat.com"
	return registry, repo, nil
}
