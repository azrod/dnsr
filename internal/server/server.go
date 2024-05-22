package server

import (
	"fmt"

	"github.com/azrod/dnsr/internal/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type (
	DNSHandler struct {
		RouterConf config.Config
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
	log.Info().Msgf("Received request for %s", domain)

	dr := DNSRequest{
		msg:               &msg,
		defaultDnsServers: h.RouterConf.Server.DefaultUpstream,
	}

	// Send the request to the upstream server
	// Define the upstream server to use
	if dnsServers := config.Md.Get(domain); dnsServers != nil {
		log.Debug().Msgf("Using upstream servers %v for domain %s", dnsServers, domain)
		dr.dnsServers = dnsServers
	} else {
		log.Debug().Msgf("Using default upstream servers %v for domain %s", dr.defaultDnsServers, domain)
	}

	// Forward the request
	switch dr.Forward(domain) {
	case dns.RcodeSuccess:
		msg.Answer = dr.msg.Answer
		if writeErr := w.WriteMsg(&msg); writeErr != nil {
			fmt.Println("Error writing response:", writeErr)
		}
	case dns.RcodeServerFailure:
		// If we get here, none of the default upstream servers responded
		// Send a SERVFAIL response
		msg.SetRcode(r, dns.RcodeServerFailure)
		if writeErr := w.WriteMsg(&msg); writeErr != nil {
			log.Error().Err(writeErr).Msg("Error writing response")
		}
	}

}
