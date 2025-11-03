package resolver

import (
	"log"
	"net"
	"strings"

	"github.com/dcarrillo/whatismyip/internal/metrics"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/internal/validator/uuid"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

type Resolver struct {
	handler *dns.ServeMux
	store   *cache.Cache
	domain  string
	rr      []string
	ipv4    []net.IP
	ipv6    []net.IP
}

func ensureDotSuffix(s string) string {
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

func Setup(store *cache.Cache) *Resolver {
	var ipv4, ipv6 []net.IP
	for _, ip := range setting.App.Resolver.Ipv4 {
		ipv4 = append(ipv4, net.ParseIP(ip))
	}
	for _, ip := range setting.App.Resolver.Ipv6 {
		ipv6 = append(ipv6, net.ParseIP(ip))
	}

	resolver := &Resolver{
		handler: dns.NewServeMux(),
		store:   store,
		domain:  ensureDotSuffix(setting.App.Resolver.Domain),
		rr:      setting.App.Resolver.ResourceRecords,
		ipv4:    ipv4,
		ipv6:    ipv6,
	}
	resolver.handler.HandleFunc(resolver.domain, resolver.resolve)
	resolver.handler.HandleFunc(".", resolver.blackHole)

	return resolver
}

func (rsv *Resolver) Handler() *dns.ServeMux {
	return rsv.handler
}

func (rsv *Resolver) blackHole(w dns.ResponseWriter, r *dns.Msg) {
	msg := startReply(r)
	msg.SetRcode(r, dns.RcodeRefused)
	w.WriteMsg(msg)
	logger(w, r.Question[0], msg.Rcode)
	metrics.RecordDNSQuery(dns.TypeToString[r.Question[0].Qtype], dns.RcodeToString[msg.Rcode])
}

func (rsv *Resolver) resolve(w dns.ResponseWriter, r *dns.Msg) {
	msg := startReply(r)
	q := r.Question[0]
	ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	for _, res := range rsv.rr {
		t := strings.Split(res, " ")[2]
		if q.Qtype == dns.StringToType[t] {
			brr, err := buildRR(rsv.domain + " " + res)
			if err != nil {
				msg.SetRcode(r, dns.RcodeServerFailure)
				logger(w, q, msg.Rcode, err.Error())
			} else {
				msg.Answer = append(msg.Answer, brr)
				logger(w, q, msg.Rcode)
			}
			w.WriteMsg(msg)
			metrics.RecordDNSQuery(dns.TypeToString[q.Qtype], dns.RcodeToString[msg.Rcode])
			return
		}
	}

	lowerName := strings.ToLower(q.Name) // lowercase because of dns-0x20
	subDomain := strings.Split(lowerName, ".")[0]
	switch {
	case uuid.IsValid(subDomain):
		msg.SetRcode(r, rsv.getIP(q, msg))
		rsv.store.Add(subDomain, ip, cache.DefaultExpiration)
	case lowerName == rsv.domain:
		msg.SetRcode(r, rsv.getIP(q, msg))
	default:
		msg.SetRcode(r, dns.RcodeRefused)
	}

	w.WriteMsg(msg)
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

func buildRR(rrs string) (dns.RR, error) {
	rr, err := dns.NewRR(rrs)
	if err != nil {
		return nil, err
	}

	return rr, nil
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
