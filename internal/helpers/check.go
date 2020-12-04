package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pires/go-proxyproto"

	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/pkg/errors"
)

func Check(protocol string, ip string, port uint16, path string, status int, proxyProdocol bool) (string, error) {

	ipPort := fmt.Sprintf("%s:%d", ip, port)
	if protocol == "tcp" {
		return check(checks.NewPingCheck("tcp", checks.NewDialPinger("tcp", ipPort), 2*time.Second))
	}
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	roundTripper := &http.Transport{
		TLSClientConfig: &tls.Config{
			// Insecure health checks are ok
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}

	if proxyProdocol {
		roundTripper.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {

			target := &net.TCPAddr{
				IP:   net.ParseIP(ip),
				Port: int(port),
			}

			conn, err := net.DialTCP("tcp", nil, target)
			if err != nil {
				return nil, err
			}

			header := &proxyproto.Header{
				Version:           1,
				Command:           proxyproto.PROXY,
				TransportProtocol: proxyproto.TCPv4,
				SourceAddr: &net.TCPAddr{
					IP:   net.ParseIP("10.1.1.1"),
					Port: 1000,
				},
				DestinationAddr: target,
			}
			if _, err := header.WriteTo(conn); err != nil {
				return nil, err
			}

			return conn, nil
		}
	}

	return check(checks.NewHTTPCheck(checks.HTTPCheckConfig{
		CheckName:      "http",
		Timeout:        1 * time.Second,
		URL:            fmt.Sprintf("%s://%s%s", protocol, ipPort, path),
		ExpectedStatus: status,
		Options: []checks.RequestOption{func(r *http.Request) {
			r.Close = true
		}},
		Client: &http.Client{
			Transport: roundTripper,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}))
}

func check(check checks.Check, err error) (string, error) {
	msg, err := checks.Must(check, err).Execute()
	message, ok := msg.(string)
	if !ok {
		return "", err
	}
	return message, errors.Wrap(err, fmt.Sprintf("%s", message))
}
