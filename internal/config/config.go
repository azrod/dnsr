// Config is a package that holds the configuration for the application.
// Define the DNS upstream server to use.

package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	Cfg   = &Config{}
	Md    = make(MatchDomains)
	mutex = &sync.RWMutex{}
)

type (
	MatchDomains map[*regexp.Regexp][]string

	Upstream struct {
		Name       string           `yaml:"name"`
		DNSServers []string         `yaml:"servers"`
		HostRegex  []string         `yaml:"regex"`
		Regex      []*regexp.Regexp `yaml:"-"`
	}

	Server struct {
		Host            string   `yaml:"host"`
		Port            int      `yaml:"port"`
		DefaultUpstream []string `yaml:"default_upstream"`
	}

	Config struct {
		Server    Server     `yaml:"server"`
		Upstreams []Upstream `yaml:"upstreams"`
		// ExternalUpstreams is a list of URLs to fetch the upstreams from.
		ExternalUpstreams []ExternalUpstreamConfig `yaml:"external_upstreams"`
	}

	ExternalUpstreamConfig struct {
		URL      string `yaml:"url"`
		Token    string `yaml:"token"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	ExternalUpstream struct {
		Upstreams []Upstream `yaml:"upstreams"`
	}
)

func (u *Upstream) CompileRegex() {
	u.Regex = make([]*regexp.Regexp, len(u.HostRegex))
	for i, r := range u.HostRegex {
		u.Regex[i] = regexp.MustCompile(r)
	}
}

func (u *Upstream) CompileDnsServers() {
	for i, server := range u.DNSServers {
		u.DNSServers[i] = net.JoinHostPort(string(server), "53")
	}
}

func (m *MatchDomains) Add(regex *regexp.Regexp, servers []string) {
	(*m)[regex] = servers
}

func (m *MatchDomains) Get(domain string) []string {
	for regex, servers := range *m {
		if regex.MatchString(domain) {
			return servers
		}
	}
	return nil
}

// Clear clears the MatchDomains map.
func (m *MatchDomains) Clear() {
	*m = make(map[*regexp.Regexp][]string)
}

// Compute MatchDomains from the Upstreams
func (m *MatchDomains) ComputeMatchDomains() {
	mutex.Lock()
	defer mutex.Unlock()
	m.Clear()
	for _, u := range Cfg.Upstreams {
		for _, r := range u.Regex {
			m.Add(r, u.DNSServers)
		}
	}
}

// GetListenAddress returns the address to listen on.
func (s *Server) GetListenAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// ReadConfig reads the configuration from the given file.
func ReadConfig(file string) error {
	// Open the file
	osFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer osFile.Close()

	mutex.Lock()
	defer mutex.Unlock()

	// Decode the file
	if err := yaml.NewDecoder(osFile).Decode(Cfg); err != nil {
		return err
	}

	for i, server := range Cfg.Server.DefaultUpstream {
		Cfg.Server.DefaultUpstream[i] = net.JoinHostPort(server, "53")
	}

	// Compile the regex
	// TODO Parallelize this
	for i := range Cfg.Upstreams {
		Cfg.Upstreams[i].CompileRegex()
		Cfg.Upstreams[i].CompileDnsServers()
	}

	return nil
}

// Watch will watch the configuration file for changes.
func WatchConfigFile(file string, done chan bool) {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					log.Info().Msg("Config file changed. Loading new configuration.")
					if err := ReadConfig(file); err != nil {
						log.Error().Err(err).Msg("Failed to read config file.")
					}
				}
				// watch for errors
			case err := <-watcher.Errors:
				log.Error().Err(err).Msg("Error watching config file.")
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(file); err != nil {
		fmt.Println("ERROR", err)
	}

	<-done
}
