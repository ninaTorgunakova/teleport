// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reversetunnelclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/sshutils"
	alpncommon "github.com/gravitational/teleport/lib/srv/alpnproxy/common"
	"github.com/gravitational/teleport/lib/utils/proxy"
)

// NewTunnelAuthDialer creates a new instance of TunnelAuthDialer
func NewTunnelAuthDialer(config TunnelAuthDialerConfig) (*TunnelAuthDialer, error) {
	if err := config.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return &TunnelAuthDialer{
		TunnelAuthDialerConfig: config,
	}, nil
}

// TunnelAuthDialerConfig specifies TunnelAuthDialer configuration.
type TunnelAuthDialerConfig struct {
	// Resolver retrieves the address of the proxy
	Resolver Resolver
	// ClientConfig is SSH tunnel client config
	ClientConfig *ssh.ClientConfig
	// Log is used for logging.
	Log logrus.FieldLogger
	// InsecureSkipTLSVerify is whether to skip certificate validation.
	InsecureSkipTLSVerify bool
	// ClusterCAs contains cluster CAs.
	ClusterCAs *x509.CertPool
}

func (c *TunnelAuthDialerConfig) CheckAndSetDefaults() error {
	if c.Resolver == nil {
		return trace.BadParameter("missing tunnel address resolver")
	}
	if c.ClusterCAs == nil {
		return trace.BadParameter("missing cluster CAs")
	}
	return nil
}

// TunnelAuthDialer connects to the Auth Server through the reverse tunnel.
type TunnelAuthDialer struct {
	// TunnelAuthDialerConfig is the TunnelAuthDialer configuration.
	TunnelAuthDialerConfig
}

// DialContext dials auth server via SSH tunnel
func (t *TunnelAuthDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	// Connect to the reverse tunnel server.
	opts := []proxy.DialerOptionFunc{
		proxy.WithInsecureSkipTLSVerify(t.InsecureSkipTLSVerify),
	}

	addr, mode, err := t.Resolver(ctx)
	if err != nil {
		t.Log.Errorf("Failed to resolve tunnel address: %v", err)
		return nil, trace.Wrap(err)
	}

	if mode == types.ProxyListenerMode_Multiplex {
		opts = append(opts, proxy.WithALPNDialer(client.ALPNDialerConfig{
			TLSConfig: &tls.Config{
				NextProtos:         []string{string(alpncommon.ProtocolReverseTunnel)},
				InsecureSkipVerify: t.InsecureSkipTLSVerify,
			},
			DialTimeout:             t.ClientConfig.Timeout,
			ALPNConnUpgradeRequired: client.IsALPNConnUpgradeRequired(ctx, addr.Addr, t.InsecureSkipTLSVerify),
			GetClusterCAs:           client.ClusterCAsFromCertPool(t.ClusterCAs),
		}))
	}

	dialer := proxy.DialerFromEnvironment(addr.Addr, opts...)
	sconn, err := dialer.Dial(ctx, addr.AddrNetwork, addr.Addr, t.ClientConfig)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// Build a net.Conn over the tunnel. Make this an exclusive connection:
	// close the net.Conn as well as the channel upon close.
	conn, _, err := sshutils.ConnectProxyTransport(
		sconn.Conn,
		&sshutils.DialReq{
			Address: RemoteAuthServer,
		},
		true,
	)
	if err != nil {
		return nil, trace.NewAggregate(err, sconn.Close())
	}
	return conn, nil
}
