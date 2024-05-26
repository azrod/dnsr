package server

import (
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// Forwards DNS requests to the appropriate upstream server.
func (d *DNSRequest) Forward(domain string) (dnsRCode int) {
	// Create a new client
	c := new(dns.Client)

	dnsServers := d.dnsServers
	dnsServers = append(dnsServers, d.defaultDNSServers...)

	m := new(dns.Msg)
	m.RecursionDesired = true

	// Send the request to the DNS servers
	for _, dnsServer := range dnsServers {
		m.SetQuestion(dns.Fqdn(domain), d.msg.Question[0].Qtype)
		upstreamResponse, timeD, err := c.Exchange(m, dnsServer)
		if err != nil {
			log.Error().Msgf("Error getting upstream response: %v", err)
			continue
		}
		if upstreamResponse.Rcode == dns.RcodeSuccess {
			log.Info().Msgf("Sending request to %s for %s took %v", dnsServer, domain, timeD)
			d.msg.Answer = upstreamResponse.Answer
			return dns.RcodeSuccess
		}
	}

	// If we get here, we didn't get a response from any of the upstream servers
	// Return a SERVFAIL
	return dns.RcodeServerFailure
}
