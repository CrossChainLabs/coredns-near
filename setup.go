package example

import (
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"
)

// init registers this plugin.
func init() { plugin.Register("example", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	connection, nearLinkNameServers, ipfsGatewayAs, ipfsGatewayAAAAs, err := nearParse(c)

	fmt.Println("connection", connection)
	fmt.Println("nearLinkNameServers", nearLinkNameServers)
	fmt.Println("ipfsGatewayAs", ipfsGatewayAs)
	fmt.Println("ipfsGatewayAAAAs", ipfsGatewayAAAAs)
	fmt.Println("err", err)

	c.Next() // Ignore "example" and give us the next token.
	if c.NextArg() {
		// If there was another token, return an error, because we don't have any configuration.
		// Any errors returned from this setup function should be wrapped with plugin.Error, so we
		// can present a slightly nicer error message to the user.
		return plugin.Error("example", c.ArgErr())
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Example{Next: next}
	})

	// All OK, return a nil error.
	return nil
}

func nearParse(c *caddy.Controller) (string, []string, []string, []string, error) {
	var connection string
	nearLinkNameServers := make([]string, 0)
	ipfsGatewayAs := make([]string, 0)
	ipfsGatewayAAAAs := make([]string, 0)

	c.Next()
	for c.NextBlock() {
		switch strings.ToLower(c.Val()) {
		case "connection":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", nil, nil, nil, c.Errf("invalid connection; no value")
			}
			if len(args) > 1 {
				return "", nil, nil, nil, c.Errf("invalid connection; multiple values")
			}
			connection = args[0]
		case "nearlinknameservers":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", nil, nil, nil, c.Errf("invalid nearlinknameservers; no value")
			}
			nearLinkNameServers = make([]string, len(args))
			copy(nearLinkNameServers, args)
		case "ipfsgatewaya":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", nil, nil, nil, c.Errf("invalid IPFS gateway A; no value")
			}
			ipfsGatewayAs = make([]string, len(args))
			copy(ipfsGatewayAs, args)
		case "ipfsgatewayaaaa":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", nil, nil, nil, c.Errf("invalid IPFS gateway AAAA; no value")
			}
			ipfsGatewayAAAAs = make([]string, len(args))
			copy(ipfsGatewayAAAAs, args)
		default:
			return "", nil, nil, nil, c.Errf("unknown value %v", c.Val())
		}
	}
	if connection == "" {
		return "", nil, nil, nil, c.Errf("no connection")
	}
	if len(nearLinkNameServers) == 0 {
		return "", nil, nil, nil, c.Errf("no nearlinknameservers")
	}
	for i := range nearLinkNameServers {
		if !strings.HasSuffix(nearLinkNameServers[i], ".") {
			nearLinkNameServers[i] = nearLinkNameServers[i] + "."
		}
	}
	return connection, nearLinkNameServers, ipfsGatewayAs, ipfsGatewayAAAAs, nil
}
