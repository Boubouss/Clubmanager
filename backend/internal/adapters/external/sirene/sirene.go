package sirene

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// validAPECodes lists the APE codes accepted for sporting associations.
var validAPECodes = map[string]bool{
	"9311Z": true,
	"9312Z": true,
	"9313Z": true,
	"9319Z": true,
}

// SireneVerifier is the port for verifying a SIREN against the public Sirene API.
type SireneVerifier interface {
	// Verify returns nil when the SIREN is valid and belongs to a sporting association.
	// If the API is unreachable it returns a wrapped error; callers should treat
	// an unreachable API as a soft failure (club created as pending, validated later).
	Verify(ctx context.Context, siren string) error
}

type httpVerifier struct {
	client *http.Client
}

// NewHTTPVerifier returns a SireneVerifier backed by the French public Entreprises API.
func NewHTTPVerifier() SireneVerifier {
	return &httpVerifier{
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

type sireneResult struct {
	Results []struct {
		MatchingEtablissements []struct {
			APECode string `json:"activite_principale"`
		} `json:"matching_etablissements"`
	} `json:"results"`
	Total int `json:"total_results"`
}

func (v *httpVerifier) Verify(ctx context.Context, siren string) error {
	url := fmt.Sprintf("https://recherche-entreprises.api.gouv.fr/search?q=%s&nombre=1", siren)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("sirene: build request: %w", err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("sirene: api unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sirene: unexpected status %d", resp.StatusCode)
	}

	var result sireneResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("sirene: decode response: %w", err)
	}

	if result.Total == 0 || len(result.Results) == 0 {
		return fmt.Errorf("sirene: SIREN not found")
	}

	for _, etab := range result.Results[0].MatchingEtablissements {
		if validAPECodes[etab.APECode] {
			return nil
		}
	}

	return fmt.Errorf("sirene: APE code does not match a sporting association")
}
