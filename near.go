// Package near is a CoreDNS plugin that prints "near" to stdout on every packet received.
//
// It serves as an near CoreDNS plugin with numerous code comments.
package near

import (
	"context"
	"fmt"
	nearclient "github.com/CrossChainLabs/go-nearclient"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// NEAR is an near plugin to show how to write a plugin.
type NEAR struct {
	Next                plugin.Handler
	Client              *nearclient.Client
	NEARLinkNameServers []string
	IPFSGatewayAs       []string
	IPFSGatewayAAAAs    []string
}

// ServeDNS implements the plugin.Handler interface. This method gets called when near is used
// in a Server.
func (e NEAR) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("near 1")

	// Call next plugin (if any).
	return dns.RcodeServerFailure, nil
}

// Name implements the Handler interface.
func (e NEAR) Name() string { return "near" }

func (e NEAR) Ready() bool { return true }
