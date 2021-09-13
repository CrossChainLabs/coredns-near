package near

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	b64 "encoding/base64"

	nearclient "github.com/CrossChainLabs/go-nearclient"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/labstack/gommon/log"
	"github.com/miekg/dns"
)

var emptyContentHash = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

type NEAR struct {
	Next                plugin.Handler
	Client              *nearclient.Client
	NEARDNS             string
	NEARLinkNameServers []string
	IPFSGatewayAs       []string
	IPFSGatewayAAAAs    []string
}

func (n NEAR) IsAuthoritative(domain string) bool {
	return true
}

func (n NEAR) HasRecords(domain string, name string) (bool, error) {
	return true, nil
}

func (n NEAR) Query(domain string, name string, qtype uint16, do bool) ([]dns.RR, error) {
	results := make([]dns.RR, 0)

	var contentHash []byte
	hasContentHash := false
	var err error
	if qtype == dns.TypeSOA ||
		qtype == dns.TypeNS ||
		qtype == dns.TypeTXT ||
		qtype == dns.TypeA ||
		qtype == dns.TypeAAAA {
		contentHash, err = n.obtainContentHash(name, domain)
		hasContentHash = err == nil && bytes.Compare(contentHash, emptyContentHash) > 0
	}
	if hasContentHash {
		switch qtype {
		case dns.TypeSOA:
			results, err = n.handleSOA(name, domain, contentHash)
		case dns.TypeNS:
			results, err = n.handleNS(name, domain, contentHash)
		case dns.TypeTXT:
			results, err = n.handleTXT(name, domain, contentHash)
		case dns.TypeA:
			results, err = n.handleA(name, domain, contentHash)
		case dns.TypeAAAA:
			results, err = n.handleAAAA(name, domain, contentHash)
		}
	}

	return results, err
}

func (n NEAR) handleSOA(name string, domain string, contentHash []byte) ([]dns.RR, error) {
	results := make([]dns.RR, 0)
	if len(n.NEARLinkNameServers) > 0 {
		// Create a synthetic SOA record
		now := time.Now()
		ser := ((now.Hour()*3600 + now.Minute()) * 100) / 86400
		dateStr := fmt.Sprintf("%04d%02d%02d%02d", now.Year(), now.Month(), now.Day(), ser)
		result, err := dns.NewRR(fmt.Sprintf("%s 10800 IN SOA %s hostmaster.%s %s 3600 600 1209600 300", n.NEARLinkNameServers[0], name, name, dateStr))
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (n NEAR) handleNS(name string, domain string, contentHash []byte) ([]dns.RR, error) {
	results := make([]dns.RR, 0)
	for _, nameserver := range n.NEARLinkNameServers {
		result, err := dns.NewRR(fmt.Sprintf("%s 3600 IN NS %s", domain, nameserver))
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (n NEAR) handleTXT(name string, domain string, contentHash []byte) ([]dns.RR, error) {
	results := make([]dns.RR, 0)
	txtRRSet, err := n.obtainTXTRRSet(name, domain)
	if err == nil && len(txtRRSet) != 0 {
		// We have a TXT rrset; use it
		offset := 0
		for offset < len(txtRRSet) {
			var result dns.RR
			result, offset, err = dns.UnpackRR(txtRRSet, offset)
			if err == nil {
				results = append(results, result)
			}
		}
	}

	result, err := dns.NewRR(fmt.Sprintf("%s 3600 IN TXT \"contenthash=0x%s\"", name, contentHash))
	if err != nil {
		return results, err
	}
	results = append(results, result)

	return results, nil
}

func (n NEAR) handleA(name string, domain string, contentHash []byte) ([]dns.RR, error) {
	results := make([]dns.RR, 0)

	aRRSet, err := n.obtainARRSet(name, domain)
	if err == nil && len(aRRSet) != 0 {
		// We have an A rrset; use it
		offset := 0
		for offset < len(aRRSet) {
			var result dns.RR
			result, offset, err = dns.UnpackRR(aRRSet, offset)
			if err == nil {
				results = append(results, result)
			}
		}
	} else {
		// We have a content hash but no A record; use the default
		for i := range n.IPFSGatewayAs {
			result, err := dns.NewRR(fmt.Sprintf("%s 3600 IN A %s", name, n.IPFSGatewayAs[i]))
			if err != nil {
				return results, err
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (n NEAR) handleAAAA(name string, domain string, contentHash []byte) ([]dns.RR, error) {
	results := make([]dns.RR, 0)

	aaaaRRSet, err := n.obtainAAAARRSet(name, domain)
	if err == nil && len(aaaaRRSet) != 0 {
		// We have an AAAA rrset; use it
		offset := 0
		for offset < len(aaaaRRSet) {
			var result dns.RR
			result, offset, err = dns.UnpackRR(aaaaRRSet, offset)
			if err == nil {
				results = append(results, result)
			}
		}
	} else {
		// We have a content hash but no AAAA record; use the default
		for i := range n.IPFSGatewayAAAAs {
			result, err := dns.NewRR(fmt.Sprintf("%s 3600 IN AAAA %s", name, n.IPFSGatewayAAAAs[i]))
			if err != nil {
				log.Warnf("error creating %s AAAA RR: %v", name, err)
			}
			results = append(results, result)
		}
	}
	return results, nil
}

// ServeDNS implements the plugin.Handler interface.
func (n NEAR) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Compress = true
	a.Authoritative = true
	var result Result
	a.Answer, a.Ns, a.Extra, result = Lookup(n, state)
	switch result {
	case Success:
		state.SizeAndDo(a)
		w.WriteMsg(a)
		return dns.RcodeSuccess, nil
	case NoData:
		if n.Next == nil {
			state.SizeAndDo(a)
			w.WriteMsg(a)
			return dns.RcodeSuccess, nil
		}
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	case NameError:
		a.Rcode = dns.RcodeNameError
	case ServerFailure:
		return dns.RcodeServerFailure, nil
	}
	// Unknown result...
	return dns.RcodeServerFailure, nil

}

func (n NEAR) obtainARRSet(name string, domain string) ([]byte, error) {
	nearDomain := strings.TrimSuffix(domain, ".near.")
	params := "{\"account_id\": \"" + nearDomain + "\"}"
	paramsEnc := b64.StdEncoding.EncodeToString([]byte(params))

	resp, err := n.Client.FunctionCall(n.NEARDNS, "get_a", paramsEnc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var byte_result []byte

	if err := json.Unmarshal(resp.Result, &byte_result); err != nil {
		log.Error(err)
		return nil, err
	}

	dec := string(byte_result)
	dec = strings.TrimPrefix(dec, "\"")
	dec = strings.TrimSuffix(dec, "\"")

	return []byte(dec), nil
}

func (n NEAR) obtainAAAARRSet(name string, domain string) ([]byte, error) {
	nearDomain := strings.TrimSuffix(domain, ".near.")
	params := "{\"account_id\": \"" + nearDomain + "\"}"
	paramsEnc := b64.StdEncoding.EncodeToString([]byte(params))

	resp, err := n.Client.FunctionCall(n.NEARDNS, "get_aaaa", paramsEnc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var byte_result []byte

	if err := json.Unmarshal(resp.Result, &byte_result); err != nil {
		log.Error(err)
		return nil, err
	}

	dec := string(byte_result)
	dec = strings.TrimPrefix(dec, "\"")
	dec = strings.TrimSuffix(dec, "\"")

	return []byte(dec), nil
}

func (n NEAR) obtainContentHash(name string, domain string) ([]byte, error) {
	nearDomain := strings.TrimSuffix(domain, ".near.")
	params := "{\"account_id\": \"" + nearDomain + "\"}"
	paramsEnc := b64.StdEncoding.EncodeToString([]byte(params))

	resp, err := n.Client.FunctionCall(n.NEARDNS, "get_content_hash", paramsEnc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var byte_result []byte

	if err := json.Unmarshal(resp.Result, &byte_result); err != nil {
		log.Error(err)
		return nil, err
	}

	dec := string(byte_result)
	dec = strings.TrimPrefix(dec, "\"")
	dec = strings.TrimSuffix(dec, "\"")

	return []byte(dec), nil
}

func (n NEAR) obtainTXTRRSet(name string, domain string) ([]byte, error) {
	nearDomain := strings.TrimSuffix(domain, ".near.")
	params := "{\"account_id\": \"" + nearDomain + "\"}"
	paramsEnc := b64.StdEncoding.EncodeToString([]byte(params))

	resp, err := n.Client.FunctionCall(n.NEARDNS, "get_txt", paramsEnc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var byte_result []byte

	if err := json.Unmarshal(resp.Result, &byte_result); err != nil {
		log.Error(err)
		return nil, err
	}

	dec := string(byte_result)
	dec = strings.TrimPrefix(dec, "\"")
	dec = strings.TrimSuffix(dec, "\"")

	return []byte(dec), nil
}

// Name implements the Handler interface.
func (n NEAR) Name() string { return "near" }

func (n NEAR) Ready() bool { return true }
