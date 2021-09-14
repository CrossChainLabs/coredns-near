package near

import (
	"strings"

	nearclient "github.com/CrossChainLabs/near-api-go"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("near", setup) }

// setup is the function that gets called when the config parser see the token "near". Setup is responsible
// for parsing any extra options the near plugin may have.
func setup(c *caddy.Controller) error {
	connection, nearDns, nearLinkNameServers, ipfsGatewayAs, ipfsGatewayAAAAs, err := nearParse(c)

	if err != nil {
		return plugin.Error("near", err)
	}

	client := nearclient.Client{URL: connection}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return NEAR{
			Next:                next,
			Client:              &client,
			NEARDNS:             nearDns,
			NEARLinkNameServers: nearLinkNameServers,
			IPFSGatewayAs:       ipfsGatewayAs,
			IPFSGatewayAAAAs:    ipfsGatewayAAAAs,
		}
	})

	// All OK, return a nil error.
	return nil
}

func nearParse(c *caddy.Controller) (string, string, []string, []string, []string, error) {
	var connection string
	var neardns string
	nearLinkNameServers := make([]string, 0)
	ipfsGatewayAs := make([]string, 0)
	ipfsGatewayAAAAs := make([]string, 0)

	c.Next()
	for c.NextBlock() {
		switch strings.ToLower(c.Val()) {
		case "connection":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", "", nil, nil, nil, c.Errf("invalid connection; no value")
			}
			if len(args) > 1 {
				return "", "", nil, nil, nil, c.Errf("invalid connection; multiple values")
			}
			connection = args[0]
		case "neardns":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", "", nil, nil, nil, c.Errf("invalid neardns; no value")
			}
			neardns = args[0]
		case "nearlinknameservers":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", "", nil, nil, nil, c.Errf("invalid nearlinknameservers; no value")
			}
			nearLinkNameServers = make([]string, len(args))
			copy(nearLinkNameServers, args)
		case "ipfsgatewaya":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", "", nil, nil, nil, c.Errf("invalid IPFS gateway A; no value")
			}
			ipfsGatewayAs = make([]string, len(args))
			copy(ipfsGatewayAs, args)
		case "ipfsgatewayaaaa":
			args := c.RemainingArgs()
			if len(args) == 0 {
				return "", "", nil, nil, nil, c.Errf("invalid IPFS gateway AAAA; no value")
			}
			ipfsGatewayAAAAs = make([]string, len(args))
			copy(ipfsGatewayAAAAs, args)
		default:
			return "", "", nil, nil, nil, c.Errf("unknown value %v", c.Val())
		}
	}
	if connection == "" {
		return "", "", nil, nil, nil, c.Errf("no connection")
	}
	if len(nearLinkNameServers) == 0 {
		return "", "", nil, nil, nil, c.Errf("no nearlinknameservers")
	}
	for i := range nearLinkNameServers {
		if !strings.HasSuffix(nearLinkNameServers[i], ".") {
			nearLinkNameServers[i] = nearLinkNameServers[i] + "."
		}
	}
	return connection, neardns, nearLinkNameServers, ipfsGatewayAs, ipfsGatewayAAAAs, nil
}
