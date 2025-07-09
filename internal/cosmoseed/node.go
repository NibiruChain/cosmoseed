package cosmoseed

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/cometbft/cometbft/v2/config"
	"github.com/cometbft/cometbft/v2/libs/log"
	"github.com/cometbft/cometbft/v2/p2p"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
	"github.com/cometbft/cometbft/v2/p2p/pex"
	"github.com/cometbft/cometbft/v2/p2p/transport/tcp"
	tcpconn "github.com/cometbft/cometbft/v2/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/v2/version"

	"github.com/NibiruChain/cosmoseed/internal/seedreactor"
)

type Seeder struct {
	home   string
	key    *p2p.NodeKey
	cfg    *Config
	logger log.Logger

	transport *tcp.MultiplexTransport
	book      p2p.AddrBook
	sw        *p2p.Switch
}

func NewSeeder(home string, config *Config) (*Seeder, error) {
	logOpt, err := log.AllowLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log options: %w", err)
	}
	logger := log.NewFilter(log.NewLogger(os.Stdout), logOpt)

	nodeKeyPath := path.Join(home, config.NodeKeyPath)
	addrBookPath := path.Join(home, config.AddrBookPath)

	if err := ensurePath(nodeKeyPath); err != nil {
		return nil, err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyPath)
	if err != nil {
		return nil, err
	}

	// Transport
	p2pConfig := generateP2PConfig(home, config)
	transport := createTransport(nodeKey, p2pConfig)

	// Address book
	book := pex.NewAddrBook(addrBookPath, config.AddrBookStrict)
	book.SetLogger(logger)

	// PEX Reactor
	pexReactor := seedreactor.NewReactor(
		book,
		splitAndTrimEmpty(p2pConfig.Seeds, ",", " "),
		config.PeerQueueSize,
		config.DialWorkers,
	)
	pexReactor.SetLogger(logger)

	// p2p switch
	sw := p2p.NewSwitch(p2pConfig, transport)
	sw.SetNodeKey(nodeKey)
	sw.SetLogger(logger)
	sw.SetAddrBook(book)
	sw.AddReactor("pex", pexReactor)
	nodeInfo := generateNodeInfo(nodeKey, config)
	sw.SetNodeInfo(nodeInfo)

	return &Seeder{
		home:      home,
		cfg:       config,
		logger:    logger,
		key:       nodeKey,
		transport: transport,
		book:      book,
		sw:        sw,
	}, nil
}

func (s *Seeder) Start() error {
	if err := s.cfg.Validate(); err != nil {
		return err
	}

	s.logger.Info("cosmoseed",
		"version", Version,
		"key", s.key.ID(),
		"listen", s.cfg.ListenAddr,
		"chain", s.cfg.ChainID,
		"log-level", s.cfg.LogLevel,
		"strict-routing", s.cfg.AddrBookStrict,
		"max-inbound", s.cfg.MaxInboundPeers,
		"max-outbound", s.cfg.MaxOutboundPeers,
		"max-packet-msg-payload-size", s.cfg.MaxPacketMsgPayloadSize,
		"dial-workers", s.cfg.DialWorkers,
		"peer-queue-size", s.cfg.PeerQueueSize,
	)

	addr, err := na.NewFromString(na.IDAddrString(s.key.ID(), s.cfg.ListenAddr))
	if err != nil {
		return err
	}

	if err = s.transport.Listen(*addr); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		s.logger.Info("shutting down...")
		if err := s.Stop(); err != nil {
			panic(err)
		}
	}()

	if err = s.sw.Start(); err != nil {
		return err
	}

	s.sw.Wait()
	return nil
}

func (s *Seeder) Stop() error {
	s.book.Save()
	return s.sw.Stop()
}

func (s *Seeder) GetNodeID() string {
	return s.key.ID()
}

func generateP2PConfig(home string, cfg *Config) *config.P2PConfig {
	p2pConfig := config.DefaultP2PConfig()

	p2pConfig.AddrBook = path.Join(home, cfg.AddrBookPath)
	p2pConfig.AddrBookStrict = cfg.AddrBookStrict
	p2pConfig.Seeds = cfg.Seeds
	p2pConfig.ListenAddress = cfg.ListenAddr
	p2pConfig.AllowDuplicateIP = true
	p2pConfig.MaxNumInboundPeers = cfg.MaxInboundPeers
	p2pConfig.MaxNumOutboundPeers = cfg.MaxOutboundPeers
	p2pConfig.MaxPacketMsgPayloadSize = cfg.MaxPacketMsgPayloadSize

	return p2pConfig
}

func createTransport(key *p2p.NodeKey, p2pConfig *config.P2PConfig) *tcp.MultiplexTransport {
	tcpConfig := tcpconn.DefaultMConnConfig()
	tcpConfig.FlushThrottle = p2pConfig.FlushThrottleTimeout
	tcpConfig.SendRate = p2pConfig.SendRate
	tcpConfig.RecvRate = p2pConfig.RecvRate
	tcpConfig.MaxPacketMsgPayloadSize = p2pConfig.MaxPacketMsgPayloadSize
	tcpConfig.TestFuzz = p2pConfig.TestFuzz
	tcpConfig.TestFuzzConfig = p2pConfig.TestFuzzConfig

	transport := tcp.NewMultiplexTransport(*key, tcpConfig)
	tcp.MultiplexTransportMaxIncomingConnections(p2pConfig.MaxNumInboundPeers)(transport)
	return transport
}

func generateNodeInfo(key *p2p.NodeKey, cfg *Config) p2p.NodeInfoDefault {
	return p2p.NodeInfoDefault{
		ProtocolVersion: p2p.ProtocolVersion{
			P2P:   version.P2PProtocol,
			Block: version.BlockProtocol,
		},
		DefaultNodeID: key.ID(),
		Network:       cfg.ChainID,
		Version:       version.CMTSemVer,
		Channels:      []byte{pex.PexChannel},
		ListenAddr:    cfg.ListenAddr,
		Moniker:       "cosmoseed",
	}
}
