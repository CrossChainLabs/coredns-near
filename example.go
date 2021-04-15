// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package example

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
)

// Example is an example plugin to show how to write a plugin.
type Example struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (e Example) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("example 1")

	// Call next plugin (if any).
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, nil, r)
}

// Name implements the Handler interface.
func (e Example) Name() string { return "example" }
