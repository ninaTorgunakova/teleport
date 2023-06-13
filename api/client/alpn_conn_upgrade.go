/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/utils"
	"github.com/gravitational/teleport/api/utils/pingconn"
	"github.com/gravitational/teleport/api/utils/tlsutils"
)

// IsALPNConnUpgradeRequired returns true if a tunnel is required through a HTTP
// connection upgrade for ALPN connections.
//
// The function makes a test connection to the Proxy Service and checks if the
// ALPN is supported. If not, the Proxy Service is likely behind an AWS ALB or
// some custom proxy services that strip out ALPN and SNI information on the
// way to our Proxy Service.
//
// In those cases, the Teleport client should make a HTTP "upgrade" call to the
// Proxy Service to establish a tunnel for the originally planned traffic to
// preserve the ALPN and SNI information.
func IsALPNConnUpgradeRequired(ctx context.Context, addr string, insecure bool, opts ...DialOption) bool {
	if result, ok := OverwriteALPNConnUpgradeRequirementByEnv(addr); ok {
		return result
	}

	// Use NewDialer which takes care of ProxyURL, and use a shorter I/O
	// timeout to avoid blocking caller.
	baseDialer := NewDialer(
		ctx,
		defaults.DefaultIdleTimeout,
		5*time.Second,
		append(opts,
			WithInsecureSkipVerify(insecure),
			WithALPNConnUpgrade(false),
		)...,
	)

	tlsConfig := &tls.Config{
		NextProtos:         []string{string(constants.ALPNSNIProtocolReverseTunnel)},
		InsecureSkipVerify: insecure,
	}
	testConn, err := tlsutils.TLSDial(ctx, baseDialer, "tcp", addr, tlsConfig)
	if err != nil {
		if isRemoteNoALPNError(err) {
			logrus.Debugf("ALPN connection upgrade required for %q: %v. No ALPN protocol is negotiated by the server.", addr, true)
			return true
		}

		// If dialing TLS fails for any other reason, we assume connection
		// upgrade is not required so it will fallback to original connection
		// method.
		logrus.Infof("ALPN connection upgrade test failed for %q: %v.", addr, err)
		return false
	}
	defer testConn.Close()

	// Upgrade required when ALPN is not supported on the remote side so
	// NegotiatedProtocol comes back as empty.
	result := testConn.ConnectionState().NegotiatedProtocol == ""
	logrus.Debugf("ALPN connection upgrade required for %q: %v.", addr, result)
	return result
}

func isRemoteNoALPNError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "remote error" && strings.Contains(opErr.Err.Error(), "tls: no application protocol")
}

// OverwriteALPNConnUpgradeRequirementByEnv overwrites ALPN connection upgrade
// requirement by environment variable.
//
// TODO(greedy52) DELETE in 15.0
func OverwriteALPNConnUpgradeRequirementByEnv(addr string) (bool, bool) {
	envValue := os.Getenv(defaults.TLSRoutingConnUpgradeEnvVar)
	if envValue == "" {
		return false, false
	}
	result := isALPNConnUpgradeRequiredByEnv(addr, envValue)
	logrus.WithField(defaults.TLSRoutingConnUpgradeEnvVar, envValue).Debugf("ALPN connection upgrade required for %q: %v.", addr, result)
	return result, true
}

// isALPNConnUpgradeRequiredByEnv checks if ALPN connection upgrade is required
// based on provided env value.
//
// The env value should contain a list of conditions separated by either ';' or
// ','. A condition is in format of either '<addr>=<bool>' or '<bool>'. The
// former specifies the upgrade requirement for a specific address and the
// later specifies the upgrade requirement for all other addresses. By default,
// upgrade is not required if target is not specified in the env value.
//
// Sample values:
// true
// <some.cluster.com>=yes,<another.cluster.com>=no
// 0,<some.cluster.com>=1
func isALPNConnUpgradeRequiredByEnv(addr, envValue string) bool {
	tokens := strings.FieldsFunc(envValue, func(r rune) bool {
		return r == ';' || r == ','
	})

	var upgradeRequiredForAll bool
	for _, token := range tokens {
		switch {
		case strings.ContainsRune(token, '='):
			if _, boolText, ok := strings.Cut(token, addr+"="); ok {
				upgradeRequiredForAddr, err := utils.ParseBool(boolText)
				if err != nil {
					logrus.Debugf("Failed to parse %v: %v", envValue, err)
				}
				return upgradeRequiredForAddr
			}

		default:
			if boolValue, err := utils.ParseBool(token); err != nil {
				logrus.Debugf("Failed to parse %v: %v", envValue, err)
			} else {
				upgradeRequiredForAll = boolValue
			}
		}
	}
	return upgradeRequiredForAll
}

// alpnConnUpgradeDialer makes an "HTTP" upgrade call to the Proxy Service then
// tunnels the connection with this connection upgrade.
type alpnConnUpgradeDialer struct {
	dialer    ContextDialer
	tlsConfig *tls.Config
	withPing  bool
}

// newALPNConnUpgradeDialer creates a new alpnConnUpgradeDialer.
func newALPNConnUpgradeDialer(dialer ContextDialer, tlsConfig *tls.Config, withPing bool) ContextDialer {
	return &alpnConnUpgradeDialer{
		dialer:    dialer,
		tlsConfig: tlsConfig,
		withPing:  withPing,
	}
}

// DialContext implements ContextDialer
func (d *alpnConnUpgradeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	logrus.Debugf("ALPN connection upgrade for %v.", addr)

	tlsConn, err := tlsutils.TLSDial(ctx, d.dialer, network, addr, d.tlsConfig.Clone())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	upgradeURL := url.URL{
		Host:   addr,
		Scheme: "https",
		Path:   constants.WebAPIConnUpgrade,
	}

	conn, err := upgradeConnThroughWebAPI(tlsConn, upgradeURL, d.upgradeType())
	if err != nil {
		return nil, trace.NewAggregate(tlsConn.Close(), err)
	}
	return conn, nil
}

func (d *alpnConnUpgradeDialer) upgradeType() string {
	if d.withPing {
		return constants.WebAPIConnUpgradeTypeALPNPing
	}
	return constants.WebAPIConnUpgradeTypeALPN
}

func upgradeConnThroughWebAPI(conn net.Conn, api url.URL, upgradeType string) (net.Conn, error) {
	req, err := http.NewRequest(http.MethodGet, api.String(), nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	req.Header.Add(constants.WebAPIConnUpgradeHeader, upgradeType)

	// Set "Connection" header to meet RFC spec:
	// https://datatracker.ietf.org/doc/html/rfc2616#section-14.42
	// Quote: "the upgrade keyword MUST be supplied within a Connection header
	// field (section 14.10) whenever Upgrade is present in an HTTP/1.1
	// message."
	//
	// Some L7 load balancers/reverse proxies like "ngrok" and "tailscale"
	// require this header to be set to complete the upgrade flow. The header
	// must be set on both the upgrade request here and the 101 Switching
	// Protocols response from the server.
	req.Header.Add(constants.WebAPIConnUpgradeConnectionHeader, constants.WebAPIConnUpgradeConnectionType)

	// Send the request and check if upgrade is successful.
	if err = req.Write(conn); err != nil {
		return nil, trace.Wrap(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer resp.Body.Close()

	if http.StatusSwitchingProtocols != resp.StatusCode {
		if http.StatusNotFound == resp.StatusCode {
			return nil, trace.NotImplemented(
				"connection upgrade call to %q with upgrade type %v failed with status code %v. Please upgrade the server and try again.",
				constants.WebAPIConnUpgrade,
				upgradeType,
				resp.StatusCode,
			)
		}
		return nil, trace.BadParameter("failed to switch Protocols %v", resp.StatusCode)
	}

	if upgradeType == constants.WebAPIConnUpgradeTypeALPNPing {
		return pingconn.New(conn), nil
	}
	return conn, nil
}
