// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package near

import (
	"context"
	"fmt"
	nearclient "github.com/CrossChainLabs/go-nearclient"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// NEAR is an example plugin to show how to write a plugin.
type NEAR struct {
	Next                plugin.Handler
	Client              *nearclient.Client
	NEARLinkNameServers []string
	IPFSGatewayAs       []string
	IPFSGatewayAAAAs    []string
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (e NEAR) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("example 1")

	// Call next plugin (if any).
	return dns.RcodeServerFailure, nil
}

// Name implements the Handler interface.
func (e NEAR) Name() string { return "example" }

func (e NEAR) Ready() bool { return true }
