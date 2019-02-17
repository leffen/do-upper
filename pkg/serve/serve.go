package serve

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/sirupsen/logrus"
)

// Server defines to common server properties
type Server struct {
}

// Run a trace server
func (s *Server) Run(ctx context.Context, targets []string, sleepTimeSeconds int64) error {
	logrus.Infof("Running serve. Targets : %#v timeBetween: %d", targets, sleepTimeSeconds)

	ch := make(chan pingResponse)
	for _, target := range targets {
		go func(url string) {
			err := s.timeSite(ctx, url, sleepTimeSeconds, ch)
			if err != ctx.Err() {
				logrus.Errorf("Unexpected error from timeSite %s", err)
			}
		}(target)
	}

	fa := newPingResponseFileAppender("data/timings.json")

	return s.responseCollector(ctx, ch, fa)
}

func (s *Server) responseCollector(ctx context.Context, ch chan pingResponse, notifier pingResponseNotifier) error {

	err := notifier.OnStart()
	if err != nil {
		return err
	}

	defer notifier.OnEnd()

	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("Quiting RUN due to context done")
			return ctx.Err()
		case resp := <-ch:
			notifier.OnNewMeasurement(resp)
		}
	}
}

func (s *Server) timeSite(ctx context.Context, url string, sleepTime int64, ch chan pingResponse) error {
	pr := pingResponse{url: url}

	for {
		pr.clearTimers()

		pr.pingNum++
		//logrus.Debugf("%06d %-30.30s Start PING", pr.pingNum, url)
		req, _ := http.NewRequest("GET", url, nil)

		var start, connect, dns, tlsHandshake time.Time

		trace := &httptrace.ClientTrace{
			DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
			DNSDone: func(ddi httptrace.DNSDoneInfo) {
				pr.dnsDuration = time.Since(dns)
			},

			TLSHandshakeStart: func() { tlsHandshake = time.Now() },
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				pr.tlsDuration = time.Since(tlsHandshake)
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				pr.connectTime = time.Since(connect)
			},

			GotFirstResponseByte: func() {
				pr.dur = time.Since(start)

			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		start = time.Now()
		if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
			// Not to self. Check if all watchers quite then server must shutdown
			logrus.Errorf("Unable to ping %s with error : %s", url, err)
			return err
		}
		pr.totalDur = time.Since(start)
		ch <- pr

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(sleepTime)):
			//	logrus.Debug("Next check")
		}

	}

}
