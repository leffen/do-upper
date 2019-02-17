package serve

import (
	"context"
	"crypto/tls"
	"encoding/json"
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

type pingResponseNotifier interface {
	OnStart() error
	OnNewMeasurement(p pingResponse) error
	OnEnd() error
}

func (p *pingResponse) clearTimers() {
	p.dur = 0
	p.dnsDuration = 0
	p.tlsDuration = 0
	p.connectTime = 0
}

func (p *pingResponse) toJSON() (string, error) {
	resp := struct {
		PingNum       int    `json:"ping_num"`
		Time          string `json:"time"`
		URL           string `json:"url"`
		Duration      int64  `json:"duration_ns"`
		TotalDuration int64  `json:"total_dutation_ns"`
		DNSDuration   int64  `json:"dns_duration_ns"`
		TLSDuration   int64  `json:"tls_duartion_ns"`
		ConnectTime   int64  `json:"connect_time_ns"`
	}{
		p.pingNum, time.Now().Format(time.RFC3339), p.url, p.dur.Nanoseconds(), p.totalDur.Nanoseconds(), p.dnsDuration.Nanoseconds(), p.tlsDuration.Nanoseconds(), p.connectTime.Nanoseconds(),
	}

	b, err := json.Marshal(resp)
	return string(b), err
}

func (p *pingResponse) debugOut() {
	logrus.Debugf("DNS Done: %v\n", p.dnsDuration)
	logrus.Debugf("TLS Handshake: %v\n", p.tlsDuration)
	logrus.Debugf("Connect time: %v\n", p.connectTime)
	logrus.Debugf("Time from start to first byte: %v\n", p.dur)
	logrus.Debugf("Total time: %v\n", p.totalDur)
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

	fa := newPingResponseFileAppender("timings.json")

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

func (s *Server) timeSite(ctx context.Context, url string, sleepTime int, ch chan pingResponse) error {
	pr := pingResponse{url: url}

	for {
		logrus.Debugf("Pinging site %s", url)
		pr.clearTimers()
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
				pr.pingNum++
				pr.dur = time.Since(start)

			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		start = time.Now()
		if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
			return err
		}
		pr.totalDur = time.Since(start)
		ch <- pr
		pr.debugOut()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(sleepTime)):
			logrus.Debug("Next check")
		}

	}

}
