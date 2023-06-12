/*
Copyright 2023 Gravitational, Inc.

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

package reverseproxy

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// X-* Header names.
const (
	XForwardedProto  = "X-Forwarded-Proto"
	XForwardedFor    = "X-Forwarded-For"
	XForwardedHost   = "X-Forwarded-Host"
	XForwardedPort   = "X-Forwarded-Port"
	XForwardedServer = "X-Forwarded-Server"
	XRealIP          = "X-Real-Ip"
)

// XHeaders X-* headers.
var XHeaders = []string{
	XForwardedProto,
	XForwardedFor,
	XForwardedHost,
	XForwardedPort,
	XForwardedServer,
	XRealIP,
}

const (
	// ContentLength is the Content-Length header.
	ContentLength = "Content-Length"
)

// Forwarder is a reverse proxy that forwards http requests to another server.
type Forwarder struct {
	passHostHeader bool
	headerRewriter Rewriter
	*httputil.ReverseProxy
}

// New returns a new reverse proxy that forwards to the given url.
// If passHostHeader is true, the Host header will be copied from the
// request to the forwarded request. Otherwise, the Host header will be
// set to the host portion of the url.
func New(passHostHeader bool, opts ...Option) *Forwarder {
	fwd := &Forwarder{
		passHostHeader: passHostHeader,
		headerRewriter: NewHeaderRewriter(),
		ReverseProxy: &httputil.ReverseProxy{
			ErrorHandler: DefaultHandler.ServeHTTP,
		},
	}
	// Apply options.
	for _, opt := range opts {
		opt(fwd)
	}

	// Director is called by the ReverseProxy to modify the request.
	fwd.Director = func(request *http.Request) {
		modifyRequest(request)
		if fwd.headerRewriter != nil {
			fwd.headerRewriter.Rewrite(request)
		}
		if !fwd.passHostHeader {
			request.Host = request.URL.Host
		}
	}

	return fwd
}

// Option is a functional option for the forwarder.
type Option func(*Forwarder)

// WithFlushInterval sets the flush interval for the forwarder.
func WithFlushInterval(interval time.Duration) Option {
	return func(rp *Forwarder) {
		rp.FlushInterval = interval
	}
}

// WithLogger sets the logger for the forwarder. It uses the logger.Writer()
// method to get the io.Writer to use for the stdlib logger.
func WithLogger[T interface{ Writer() *io.PipeWriter }](logger T,
) Option {
	return func(rp *Forwarder) {
		// TODO: this is a hack to get the stdlib logger to work with the
		// logrus logger.
		rp.ReverseProxy.ErrorLog = log.New(logger.Writer(), "", log.LstdFlags)
	}
}

// WithRoundTripper sets the round tripper for the forwarder.
func WithRoundTripper(transport http.RoundTripper) Option {
	return func(rp *Forwarder) {
		rp.Transport = transport
	}
}

// WithErrorHandler sets the error handler for the forwarder.
func WithErrorHandler(e ErrorHandlerFunc) Option {
	return func(rp *Forwarder) {
		rp.ErrorHandler = e
	}
}

// WithRewriter sets the header rewriter for the forwarder.
func WithRewriter(h Rewriter) Option {
	return func(rp *Forwarder) {
		rp.headerRewriter = h
	}
}

// Modify the request to handle the target URL.
func modifyRequest(outReq *http.Request) {
	u := getURLFromRequest(outReq)

	outReq.URL.Path = u.Path
	outReq.URL.RawPath = u.RawPath
	outReq.URL.RawQuery = u.RawQuery
	outReq.RequestURI = "" // Outgoing request should not have RequestURI

	outReq.Proto = "HTTP/1.1"
	outReq.ProtoMajor = 1
	outReq.ProtoMinor = 1
}

// getURLFromRequest returns the URL from the request object. If the request
// RequestURI is non-empty and parsable, it will be used. Otherwise, the URL
// will be used.
func getURLFromRequest(req *http.Request) *url.URL {
	// If the Request was created by Go via a real HTTP request,
	// RequestURI will contain the original query string.
	// If the Request was created in code,
	// RequestURI will be empty, and we will use the URL object instead
	u := req.URL
	if req.RequestURI != "" {
		parsedURL, err := url.ParseRequestURI(req.RequestURI)
		if err == nil {
			return parsedURL
		}
	}
	return u
}
