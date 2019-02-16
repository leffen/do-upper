package serve

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/sirupsen/logrus"
)

// Server defines to common server properties
type Server struct {
}

type pingResponse struct {
	pingNum     int
	url         string
	dur         time.Duration
	totalDur    time.Duration
	dnsDuration time.Duration
	tlsDuration time.Duration
	connectTime time.Duration
}

func (p *pingResponse) clearTimers() {
	p.dur = 0
	p.dnsDuration = 0
	p.tlsDuration = 0
	p.connectTime = 0
}

// Run a trace server
func (s *Server) Run(ctx context.Context, targets []string, sleepTimeSeconds int) error {
	logrus.Info("Running serve")
	ch := make(chan pingResponse)
	for _, target := range targets {
		go func(url string) {
			err := s.timeSite(ctx, url, sleepTimeSeconds, ch)
			if err != ctx.Err() {
				logrus.Errorf("Unexpected error from timeSite %s", err)
			}
		}(target)
	}
	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("Quiting RUN due to context done")
			return ctx.Err()
		case resp := <-ch:
			fmt.Printf("RESP :%#v\n", resp)
		}
	}

}

func (s *Server) timeSite(ctx context.Context, url string, sleepTime int, ch chan pingResponse) error {
	pr := pingResponse{}

	for {
		pr.clearTimers()
		req, _ := http.NewRequest("GET", url, nil)

		var start, connect, dns, tlsHandshake time.Time

		trace := &httptrace.ClientTrace{
			DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
			DNSDone: func(ddi httptrace.DNSDoneInfo) {
				pr.dnsDuration = time.Since(dns)
				fmt.Printf("DNS Done: %v\n", pr.dnsDuration)
			},

			TLSHandshakeStart: func() { tlsHandshake = time.Now() },
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				pr.tlsDuration = time.Since(tlsHandshake)
				fmt.Printf("TLS Handshake: %v\n", pr.tlsDuration)
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				pr.connectTime = time.Since(connect)
				fmt.Printf("Connect time: %v\n", pr.connectTime)
			},

			GotFirstResponseByte: func() {
				pr.pingNum++
				pr.dur = time.Since(start)
				fmt.Printf("Time from start to first byte: %v\n", pr.dur)

			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		start = time.Now()
		if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
			return err
		}
		pr.totalDur = time.Since(start)
		ch <- pr
		fmt.Printf("Total time: %v\n", pr.totalDur)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * time.Duration(sleepTime)):
				break
			}
		}

	}

}