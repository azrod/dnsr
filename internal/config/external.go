package config

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type HashDBExternal struct {
	mu sync.RWMutex
	db map[string]string
}

var hdbe = HashDBExternal{
	db: make(map[string]string),
}

// LoadExternalUpstreams loads the external upstreams from the URLs provided in the configuration file
func LoadExternalUpstreams() {

	var updated = false

	client := resty.New()
	for _, url := range Cfg.ExternalUpstreams {
		log.Debug().Msgf("Fetching upstreams from %s", url.URL)

		c := client.R().
			SetHeader("Accept", "application/yaml")

		if url.Token != "" {
			c.SetAuthToken(url.Token)
		}

		if url.Username != "" && url.Password != "" {
			c.SetBasicAuth(url.Username, url.Password)
		}

		resp, err := c.Get(url.URL)
		if err != nil {
			log.Error().Err(err).Msgf("Error fetching upstreams from %s", url)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Error().Msgf("Error fetching upstreams from %s: %s", url, resp.Status())
			continue
		}

		var external ExternalUpstream

		if hdbe.HasUpdated(url.URL, resp.Body()) {
			if err := yaml.Unmarshal(resp.Body(), &external); err != nil {
				log.Error().Err(err).Msgf("Error decoding upstreams from %s", url.URL)
				continue
			}

			log.Info().Msgf("Updating upstreams from %s. Found %d upstream(s)", url.URL, len(external.Upstreams))
			for _, u := range external.Upstreams {
				u.CompileRegex()
				u.CompileDnsServers()
				mutex.Lock()
				Cfg.Upstreams = append(Cfg.Upstreams, u)
				mutex.Unlock()
			}

			updated = true
			hdbe.Update(url.URL, hdbe.ComputeHash(url.URL, resp.Body()))
		}

	}

	if updated {
		Md.ComputeMatchDomains()
	}
}

// ComputeHash computes the hash of the external upstream
func (h *HashDBExternal) ComputeHash(url string, upstreamContent []byte) string {
	return hex.EncodeToString(h.hash(upstreamContent))
}

// hash returns the hash of the external upstream
func (h *HashDBExternal) hash(content []byte) []byte {
	x := sha256.Sum256(content)
	return x[:]
}

// UpdateHash updates the hash of the external upstream
func (h *HashDBExternal) Update(url string, hash string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.db[url] = hash
}

// Exists checks if the external upstream exists
func (h *HashDBExternal) exist(url string) bool {
	_, ok := h.db[url]
	return ok
}

// HasUpdated checks if the external upstream has been updated
func (h *HashDBExternal) HasUpdated(url string, upstreamContent []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.exist(url) || h.db[url] != h.ComputeHash(url, upstreamContent)
}
