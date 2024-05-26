package config

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

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

// LoadExternalUpstreams loads the external upstreams from the URLs provided in the configuration file.
func LoadExternalUpstreams() {
	// WaitGroup to wait for all the external upstreams to be fetched
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		updated bool
	)

	for i, url := range Cfg.ExternalUpstreams {
		wg.Add(1)
		go func(url ExternalUpstreamConfig, i int) {
			defer wg.Done()

			client := resty.New()
			client.SetTimeout(5 * time.Second)
			log.Info().Msgf("Fetching upstreams (%d/%d) from %s", i+1, len(Cfg.ExternalUpstreams), url.URL)

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
				return
			}

			if resp.StatusCode() != 200 {
				log.Error().Msgf("Error fetching upstreams from %s: %s", url, resp.Status())
				return
			}

			var external ExternalUpstream

			if hdbe.HasUpdated(url.URL, resp.Body()) {
				if err := yaml.Unmarshal(resp.Body(), &external); err != nil {
					log.Error().Err(err).Msgf("Error decoding upstreams from %s", url.URL)
					return
				}

				log.Info().Msgf("Found %d upstream(s) from %s", len(external.Upstreams), url.URL)

				for _, u := range external.Upstreams {
					u.CompileRegex()
					u.CompileDNSServers()
					Cfg.mu.Lock()
					Cfg.Upstreams = append(Cfg.Upstreams, u)
					Cfg.mu.Unlock()
				}

				mu.Lock()
				updated = true
				mu.Unlock()
				hdbe.Update(url.URL, hdbe.ComputeHash(resp.Body()))
			}
		}(url, i)
	}

	wg.Wait()
	if updated {
		Md.ComputeMatchDomains()
	}
}

// ComputeHash computes the hash of the external upstream.
func (h *HashDBExternal) ComputeHash(upstreamContent []byte) string {
	return hex.EncodeToString(h.hash(upstreamContent))
}

// hash returns the hash of the external upstream.
func (h *HashDBExternal) hash(content []byte) []byte {
	x := sha256.Sum256(content)
	return x[:]
}

// UpdateHash updates the hash of the external upstream.
func (h *HashDBExternal) Update(url, hash string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.db[url] = hash
}

// Exists checks if the external upstream exists.
func (h *HashDBExternal) exist(url string) bool {
	_, ok := h.db[url]
	return ok
}

// HasUpdated checks if the external upstream has been updated.
func (h *HashDBExternal) HasUpdated(url string, upstreamContent []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.exist(url) || h.db[url] != h.ComputeHash(upstreamContent)
}
