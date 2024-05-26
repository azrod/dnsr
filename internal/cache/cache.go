package cache

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/azrod/dnsr/internal/cache/base"
	"github.com/azrod/dnsr/internal/cache/memory"
	"github.com/azrod/dnsr/internal/config"
)

// New creates a new cache.
func New() (base.Cache, error) {
	c, err := memory.New()
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Cache are successfully created")

	// Load the cache from disk
	if cachedLoaded, err := LoadCache(); err == nil && len(cachedLoaded) > 0 {
		if err := c.Load(cachedLoaded); err != nil {
			log.Error().Err(err).Msg("Error loading cache from disk")
		}
		log.Info().Msgf("Restore cache from disk (%d entries)", c.Len())
	} else {
		log.Error().Err(err).Msg("Error loading cache from disk")
	}

	go func() {
		// new ticker
		ticker := time.NewTicker(5 * time.Minute)
		tickerPrint := time.NewTicker(2 * time.Minute)
		for {
			select {
			case <-ticker.C:
				if err := PersistCache(c); err != nil {
					log.Error().Err(err).Msg("Error persisting cache")
				}
			case <-tickerPrint.C:
				log.Info().Msgf("Cache size: %d", c.Len())
			}
		}
	}()

	return c, nil
}

func registerGobTypes() {
	gob.Register(&dns.A{})
	gob.Register(&dns.AAAA{})
	gob.Register(&dns.NS{})
	gob.Register(&dns.MD{})
	gob.Register(&dns.MF{})
	gob.Register(&dns.CNAME{})
	gob.Register(&dns.SOA{})
	gob.Register(&dns.MB{})
	gob.Register(&dns.MG{})
	gob.Register(&dns.MR{})
	gob.Register(&dns.NULL{})
	gob.Register(&dns.PTR{})
	gob.Register(&dns.HINFO{})
	gob.Register(&dns.MINFO{})
	gob.Register(&dns.MX{})
	gob.Register(&dns.TXT{})
	gob.Register(&dns.RP{})
	gob.Register(&dns.AFSDB{})
	gob.Register(&dns.X25{})
	gob.Register(&dns.ISDN{})
	gob.Register(&dns.RT{})
	gob.Register(&dns.NSAPPTR{})
	gob.Register(&dns.SIG{})
	gob.Register(&dns.KEY{})
	gob.Register(&dns.PX{})
	gob.Register(&dns.GPOS{})
	gob.Register(&dns.AAAA{})
	gob.Register(&dns.LOC{})
	gob.Register(&dns.NXT{})
	gob.Register(&dns.EID{})
	gob.Register(&dns.NIMLOC{})
	gob.Register(&dns.SRV{})
	gob.Register(&dns.NAPTR{})
	gob.Register(&dns.KX{})
	gob.Register(&dns.CERT{})
	gob.Register(&dns.DNAME{})
	gob.Register(&dns.OPT{}) // EDNS
	gob.Register(&dns.APL{})
	gob.Register(&dns.DS{})
	gob.Register(&dns.SSHFP{})
	gob.Register(&dns.IPSECKEY{})
	gob.Register(&dns.RRSIG{})
	gob.Register(&dns.NSEC{})
	gob.Register(&dns.DNSKEY{})
	gob.Register(&dns.DHCID{})
	gob.Register(&dns.NSEC3{})
	gob.Register(&dns.NSEC3PARAM{})
	gob.Register(&dns.TLSA{})
	gob.Register(&dns.SMIMEA{})
	gob.Register(&dns.HIP{})
	gob.Register(&dns.NINFO{})
	gob.Register(&dns.RKEY{})
	gob.Register(&dns.TALINK{})
	gob.Register(&dns.CDS{})
	gob.Register(&dns.CDNSKEY{})
	gob.Register(&dns.OPENPGPKEY{})
	gob.Register(&dns.CSYNC{})
	gob.Register(&dns.ZONEMD{})
	gob.Register(&dns.SVCB{})
	gob.Register(&dns.HTTPS{})
	gob.Register(&dns.SPF{})
	gob.Register(&dns.UINFO{})
	gob.Register(&dns.UID{})
	gob.Register(&dns.GID{})
	gob.Register(&dns.NID{})
	gob.Register(&dns.L32{})
	gob.Register(&dns.L64{})
	gob.Register(&dns.LP{})
	gob.Register(&dns.EUI48{})
	gob.Register(&dns.EUI64{})
	gob.Register(&dns.URI{})
	gob.Register(&dns.CAA{})
	gob.Register(&dns.AVC{})
	gob.Register(&dns.AMTRELAY{})
	gob.Register(&dns.SVCBAlpn{})
	gob.Register(&dns.SVCBPort{})
	gob.Register(&dns.TA{})
	gob.Register(&dns.SVCBIPv4Hint{})
	gob.Register(&dns.SVCBIPv6Hint{})
	gob.Register(&dns.SVCBDoHPath{})
}

// PersistCache will persist the cache to disk.
func PersistCache(c base.Cache) error {
	// Read the cache from memory
	// Write the cache to disk

	if c == nil || c.Len() == 0 {
		log.Debug().Msg("No cache to persist. Ignoring")
		return nil
	}

	log.Info().Msgf("Persisting cache to disk (%d entries)", c.Len())
	values := c.GetAll()

	file, err := os.Create(getPathCache())
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	registerGobTypes()

	if err := encoder.Encode(values); err != nil {
		return fmt.Errorf("error encoding cache: %w", err)
	}

	return nil
}

// LoadCache will load the cache from disk.
func LoadCache() (map[string]base.CacheValue, error) {
	// Read the cache from disk
	file, err := os.Open(getPathCache())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Debug().Msg("Cache file does not exist")
			return nil, nil //nolint:nilnil
		default:
			return nil, fmt.Errorf("error opening cache file: %w", err)
		}
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	registerGobTypes()

	var values map[string]base.CacheValue
	return values, decoder.Decode(&values)
}

// defaultCachePath is the default path to store the cache.
const defaultCachePath = "./cache.gob"

func getPathCache() string {
	if config.Cfg.Cache.Path == "" {
		return defaultCachePath
	}

	return config.Cfg.Cache.Path
}
