package proxy

import (
	"context"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

type Resolver interface {
	LookupCNAME(targetCname string) (string, error)
}

type goNetResolver struct{}

func (goNetResolver) LookupCNAME(targetCname string) (string, error) {
	return net.LookupCNAME(targetCname)
}

func GoNetResolver() goNetResolver {
	return goNetResolver{}
}

type Monitor struct {
	targetCname string
	lastData    string
	resolver    Resolver
	interval    time.Duration
}

func NewMonitor(resolver Resolver, targetCname string, interval time.Duration) *Monitor {
	return &Monitor{
		targetCname: targetCname,
		resolver:    resolver,
		interval:    interval,
	}
}

func (m *Monitor) Run(ctx context.Context, callback func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cname, err := m.resolver.LookupCNAME(m.targetCname)
		if err != nil {
			log.Error().Err(err).Msg("failed to lookup CNAME")
			cname = m.lastData
		}
		if m.lastData != "" && m.lastData != cname {
			log.Info().Msg("CNAME changed, running callback")
			callback()
		}
		m.lastData = cname

		// Go net doesn't provide us with the TTL, therefor we poll frequently and
		// expect the appropriate levels of caching to be in place, in order for this
		// not to cause undue load on the DNS servers.
		time.Sleep(m.interval)
	}
}
