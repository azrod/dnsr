package server

import (
	"fmt"
	"strings"

	"github.com/azrod/dnsr/internal/cache/base"
	"github.com/azrod/dnsr/internal/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type (
	DNSHandler struct {
		Cache base.Cache
	}

	DNSRequest struct {
		msg               *dns.Msg
		dnsServers        []string
		defaultDnsServers []string
	}
)

// ServeDNS will handle incoming dns requests and forward them onwards
func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	domain := msg.Question[0].Name
	log.Debug().Msgf("Received request for %s", domain)

	// Add special command in domain name
	// For examples
	// clear/domain.com will clear the cache for domain.com
	// clear/all
	switch {
	case domain == "clear/all.":
		if h.Cache != nil {
			log.Info().Msg("Clearing all cache")
			if err := h.Cache.Clear(); err != nil {
				log.Error().Err(err).Msg("Error clearing cache")
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   domain,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{fmt.Sprintf("Error clearing cache: %v", err)},
				})
			} else {
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   domain,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{"Cache cleared"},
				})
			}
		} else {
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{"Cache not enabled"},
			})
		}
	case strings.HasPrefix(domain, "clear/"):
		domain = strings.TrimPrefix(domain, "clear/")
		if h.Cache != nil {
			log.Info().Msgf("Clearing cache for %s", domain)
			if err := h.Cache.Delete(domain); err != nil {
				log.Error().Err(err).Msgf("Error clearing cache for %s", domain)
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   domain,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{fmt.Sprintf("Error clearing cache: %v", err)},
				})
			} else {
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   domain,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{fmt.Sprintf("Cache cleared for %s", domain)},
				})
			}
		} else {
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{"Cache not enabled"},
			})
		}
	default:
		if h.Cache != nil && h.Cache.Exists(domain) && !h.Cache.HasExpired(domain) {
			if value, err := h.Cache.Get(domain); err == nil {
				log.Info().Msgf("Using cache for domain %s (Expire at %v)", domain, h.Cache.GetExpireAt(domain).Format("2006-01-02 15:04:05"))
				msg.Answer = append(msg.Answer, value...)
			}
		} else {

			dr := DNSRequest{
				msg:               &msg,
				defaultDnsServers: config.Cfg.Server.DefaultUpstream,
			}

			// Send the request to the upstream server
			// Define the upstream server to use
			if dnsServers := config.Md.Get(domain); dnsServers != nil {
				dr.dnsServers = dnsServers
			}

			// Forward the request
			switch dr.Forward(domain) {
			case dns.RcodeSuccess:
				msg.Answer = dr.msg.Answer
				if h.Cache != nil && len(msg.Answer) > 0 {
					log.Info().Msgf("Caching response for %s", domain)
					if err := h.Cache.Set(domain, msg.Answer); err != nil {
						log.Error().Err(err).Msg("Error writing in cache")
					}
				}
			case dns.RcodeServerFailure:
				// Special case trying get value from cache even if ttl has expired
				if h.Cache != nil && h.Cache.Exists(domain) && !h.Cache.HasExpired(domain) {
					if value, err := h.Cache.Get(domain); err == nil {
						log.Debug().Msgf("Cache hit for %s", domain)
						msg.Answer = append(msg.Answer, value...)
					}
				} else {
					// If we get here, none of the default upstream servers responded
					// Send a SERVFAIL response
					msg.SetRcode(r, dns.RcodeServerFailure)
				}
			default:
				msg.Answer = dr.msg.Answer
			}

		}
	}

	if writeErr := w.WriteMsg(&msg); writeErr != nil {
		log.Error().Err(writeErr).Msg("Error writing response")
	}
}
