package seedreactor

import (
	"fmt"
	"time"

	p2papi "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/v2/libs/log"
	"github.com/cometbft/cometbft/v2/p2p"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
	"github.com/cometbft/cometbft/v2/p2p/pex"
)

type SeedReactor struct {
	*pex.Reactor

	book        pex.AddrBook
	log         log.Logger
	addrChan    chan *AddrPair
	quitCh      chan struct{}
	dialWorkers int
	strict      bool
}

type AddrPair struct {
	Addr   *na.NetAddr
	Source *na.NetAddr
}

func NewReactor(book pex.AddrBook, seeds []string, queueSize, dialWorkers int, strict bool) *SeedReactor {
	r := pex.NewReactor(book, &pex.ReactorConfig{
		SeedMode:          true,
		Seeds:             seeds,
		EnsurePeersPeriod: 30 * time.Second,
	})

	return &SeedReactor{
		Reactor:     r,
		book:        book,
		log:         log.NewNopLogger(),
		addrChan:    make(chan *AddrPair, queueSize),
		quitCh:      make(chan struct{}),
		dialWorkers: dialWorkers,
		strict:      strict,
	}
}

func (s *SeedReactor) Start() error {
	s.StartDialWorkers(s.dialWorkers)
	return s.Reactor.Start()
}

func (s *SeedReactor) Stop() error {
	close(s.quitCh)
	return s.Reactor.Stop()
}

func (s *SeedReactor) SetLogger(logger log.Logger) {
	s.log = logger
	s.Reactor.SetLogger(logger)
}

func (s *SeedReactor) AddPeer(p p2p.Peer) {
	addr := p.SocketAddr()
	if addr == nil {
		s.log.Warn("not adding peer: no address", "id", p.ID())
		return
	}
	if s.strict && !addr.Routable() {
		s.log.Warn("not adding peer: address not routable", "id", p.ID(), "addr", addr)
		return
	}

	s.log.Info("adding/marking good peer", "id", p.ID(), "addr", addr)
	s.book.MarkGood(addr.ID)
	s.Reactor.AddPeer(p)
}

func (s *SeedReactor) Receive(e p2p.Envelope) {
	s.log.Debug("received pex message", "from", e.Src.ID(), "type", fmt.Sprintf("%T", e.Message))

	switch msg := e.Message.(type) {
	case *p2papi.PexRequest:
		s.Reactor.Receive(e)

	case *p2papi.PexAddrs:
		addrs, err := na.AddrsFromProtos(msg.Addrs)
		if err != nil {
			s.log.Error("failed to decode received addresses", "err", err)
			return
		}

		for _, addr := range addrs {
			s.log.Debug("received peer address", "addr", addr.DialString())
			select {
			case s.addrChan <- &AddrPair{
				Addr:   addr,
				Source: e.Src.SocketAddr(),
			}:
			default:
				s.log.Warn("dial queue full, dropping address", "addr", addr.DialString())
			}
		}

	default:
		s.log.Warn("received unknown PEX message type", "type", fmt.Sprintf("%T", msg))
	}
}

func (s *SeedReactor) StartDialWorkers(n int) {
	for i := 0; i < n; i++ {
		go func() {
			for {
				select {
				case addr := <-s.addrChan:
					s.log.With("dial-worker", i).Debug("dialing peer", "peer", addr)
					s.processAddr(addr)
				case <-s.quitCh:
					return
				}
			}
		}()
	}
}

func (s *SeedReactor) processAddr(addr *AddrPair) {
	if addr == nil {
		s.log.Debug("ignoring nil address")
		return
	}

	if s.strict && !addr.Addr.Routable() {
		s.log.Debug("received peer address not routable. Ignoring", "addr", addr.Addr.DialString())
		return
	}

	if s.Switch.IsDialingOrExistingAddress(addr.Addr) {
		s.log.Debug("already dialing or connected", "addr", addr)
		return
	}
	err := s.Reactor.Switch.DialPeerWithAddress(addr.Addr)
	if err != nil {
		s.log.Debug("dial failed", "addr", addr, "err", err)
		s.book.MarkAttempt(addr.Addr)
		return
	}
	s.log.Info("adding/marking good peer", "id", addr.Addr.ID, "addr", addr)
	if addr.Source == nil {
		addr.Source = s.Switch.NetAddr()
	}
	if err = s.book.AddAddress(addr.Addr, addr.Source); err != nil {
		s.log.Error("failed to add address", "addr", addr, "err", err)
		return
	}
	s.book.MarkGood(addr.Addr.ID)
}

func (s *SeedReactor) GetPeerSelection() []*na.NetAddr {
	return s.book.GetSelection()
}
