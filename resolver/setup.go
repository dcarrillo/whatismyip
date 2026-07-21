package resolver

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/dcarrillo/whatismyip/internal/metrics"
	"github.com/dcarrillo/whatismyip/internal/validator/uuid"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

type Settings struct {
	Domain          string
	ResourceRecords []string
	RedirectPort    string
	IPv4            []string
	IPv6            []string
}

type Resolver struct {
	handler *dns.ServeMux
	store   *cache.Cache
	domain  string
	rr      []dns.RR
	ipv4    []net.IP
	ipv6    []net.IP
}

func ensureDotSuffix(s string) string {
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

func Setup(store *cache.Cache, cfg Settings) (*Resolver, error) {
	domain := ensureDotSuffix(cfg.Domain)

	rr := make([]dns.RR, 0, len(cfg.ResourceRecords))
	for _, res := range cfg.ResourceRecords {
		record, err := dns.NewRR(domain + " " + res)
		if err != nil {
			return nil, fmt.Errorf("parsing resource record %q: %w", res, err)
		}
		rr = append(rr, record)
	}

	ipv4, err := parseIPs(cfg.IPv4)
	if err != nil {
		return nil, fmt.Errorf("parsing ipv4 addresses: %w", err)
	}
	ipv6, err := parseIPs(cfg.IPv6)
	if err != nil {
		return nil, fmt.Errorf("parsing ipv6 addresses: %w", err)
	}

	resolver := &Resolver{
		handler: dns.NewServeMux(),
		store:   store,
		domain:  domain,
		rr:      rr,
		ipv4:    ipv4,
		ipv6:    ipv6,
	}
	resolver.handler.HandleFunc(resolver.domain, resolver.resolve)
	resolver.handler.HandleFunc(".", resolver.blackHole)

	return resolver, nil
}

func parseIPs(addresses []string) ([]net.IP, error) {
	ips := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address %q", address)
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func (rsv *Resolver) Handler() *dns.ServeMux {
	return rsv.handler
}

func (rsv *Resolver) blackHole(w dns.ResponseWriter, r *dns.Msg) {
	msg := startReply(r)
	msg.SetRcode(r, dns.RcodeRefused)
	writeMsg(w, msg)
	logger(w, r.Question[0], msg.Rcode)
	metrics.RecordDNSQuery(dns.TypeToString[r.Question[0].Qtype], dns.RcodeToString[msg.Rcode])
}

func (rsv *Resolver) resolve(w dns.ResponseWriter, r *dns.Msg) {
	msg := startReply(r)
	q := r.Question[0]
	ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	for _, res := range rsv.rr {
		if q.Qtype == res.Header().Rrtype {
			msg.Answer = append(msg.Answer, dns.Copy(res))
			writeMsg(w, msg)
			logger(w, q, msg.Rcode)
			metrics.RecordDNSQuery(dns.TypeToString[q.Qtype], dns.RcodeToString[msg.Rcode])
			return
		}
	}

	lowerName := strings.ToLower(q.Name) // lowercase because of dns-0x20
	subDomain := strings.Split(lowerName, ".")[0]
	switch {
	case uuid.IsValid(subDomain):
		msg.SetRcode(r, rsv.getIP(q, msg))
		// Add fails when the uuid is already registered; keep the first seen
		// resolver IP for the discovery window
		_ = rsv.store.Add(subDomain, ip, cache.DefaultExpiration)
	case lowerName == rsv.domain:
		msg.SetRcode(r, rsv.getIP(q, msg))
	default:
		msg.SetRcode(r, dns.RcodeRefused)
	}

	writeMsg(w, msg)
	logger(w, q, msg.Rcode)
	metrics.RecordDNSQuery(dns.TypeToString[q.Qtype], dns.RcodeToString[msg.Rcode])
}

func (rsv *Resolver) getIP(question dns.Question, msg *dns.Msg) int {
	if question.Qtype == dns.TypeA && len(rsv.ipv4) > 0 {
		for _, ip := range rsv.ipv4 {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: setHdr(question),
				A:   ip,
			})
		}
		return dns.RcodeSuccess
	}

	if question.Qtype == dns.TypeAAAA && len(rsv.ipv6) > 0 {
		for _, ip := range rsv.ipv6 {
			msg.Answer = append(msg.Answer, &dns.AAAA{
				Hdr:  setHdr(question),
				AAAA: ip,
			})
		}
		return dns.RcodeSuccess
	}

	return dns.RcodeRefused
}

func writeMsg(w dns.ResponseWriter, msg *dns.Msg) {
	if err := w.WriteMsg(msg); err != nil {
		log.Printf("Failed to write DNS response: %s", err)
	}
}

func setHdr(q dns.Question) dns.RR_Header {
	return dns.RR_Header{
		Name:   q.Name,
		Rrtype: q.Qtype,
		Class:  dns.ClassINET,
		Ttl:    60,
	}
}

func startReply(r *dns.Msg) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	return msg
}

func logger(w dns.ResponseWriter, q dns.Question, code int, err ...string) {
	emsg := ""
	if len(err) > 0 {
		emsg = " - " + strings.Join(err, " ")
	}
	ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	log.Printf(
		"DNS %s - %s - %s - %s%s",
		ip,
		dns.TypeToString[q.Qtype],
		q.Name,
		dns.RcodeToString[code],
		emsg,
	)
}
