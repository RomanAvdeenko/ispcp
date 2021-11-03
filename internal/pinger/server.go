package pinger

import (
	"fmt"
	mynet "github.com/RomanAvdeenko/utils/net"
	"ispcp/internal/host"
	"ispcp/internal/model"
	"ispcp/internal/store"
	"ispcp/internal/store/file"
	"os"
	"runtime"
	"time"

	"net"

	"github.com/j-keck/arping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	conifg   *Config
	store    store.Store
	logger   *zerolog.Logger
	host     *host.Host
	pingChan chan model.Ping
	pongs    *model.Pongs
}

func newServer(cfg *Config, store store.Store) *Server {
	s := Server{
		conifg: cfg,
		store:  store,
		//logger: &zerolog.New(os.Stdout),
		logger:   &zerolog.Logger{},
		host:     host.NewHost(),
		pingChan: make(chan model.Ping, jobChanLen),
		pongs:    model.NewPongs(),
	}
	s.configure()
	return &s
}

func Start(cfg *Config) error {
	f, err := os.OpenFile("./store.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	store := file.New(f)
	s := newServer(cfg, store)

	refreshInterval := time.Duration(s.conifg.RestartInterval) * time.Second
	refreshTicker := time.NewTicker(refreshInterval)

	s.logger.Info().Msg(fmt.Sprintf("Start pinger with %v threads, refresh interval: %s...", s.conifg.ThreadsNumber, refreshInterval))

	go func() {
		// Start working instantly
		s.addWork()
		s.startWorkers()

		for {
			select {
			case <-refreshTicker.C:
				s.store.Store(s.pongs)
				s.addWork()
				s.startWorkers()
			}
		}
	}()

	quit := make(chan struct{})
	<-quit

	return nil
}

func (s *Server) configure() error {
	s.configureLogger()

	s.host.SetExcludeIfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.SetExcludeNetworkIPs(s.conifg.ExcludeNetIPs)
	s.host.SetLogger(s.logger)
	s.host.Configure()

	return nil
}

func (s *Server) configureLogger() {
	*s.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli})
}

// Adds work to ping all required host interfaces
func (s *Server) addWork() error {
	s.logger.Info().Msg("Starting to add work.")
	//Let's walk through the interfaces
	for _, iface := range s.host.ProcessedIfaces {
		ifaceAddrs, err := s.host.GetIfaceAddrs(iface)
		if err != nil {
			return err
		}
		for _, ifaceAddr := range ifaceAddrs {
			// Add job to workers
			go func(iface net.Interface, addr string, pingChan chan<- model.Ping) {
				s.logger.Info().Msg(fmt.Sprintf("Processed interface: %v, processed network: %v", iface.Name, ifaceAddr))
				ips, _ := mynet.GetHostsIP(addr)
				for _, ip := range ips {
					pingChan <- model.Ping{IP: ip, Iface: iface}
				}
			}(iface, ifaceAddr, s.pingChan)
		}
	}
	return nil
}

func (s *Server) startWorkers() {
	for i := 0; i < s.conifg.ThreadsNumber; i++ {
		// Start workers
		go func(pingChan <-chan model.Ping, num int) {
			for ping := range pingChan {
				ip := ping.IP
				iface := ping.Iface

				macAddr, duration, err := arping.PingOverIface(ip, iface)
				if err != nil {
					//s.logger.Debug().Msg(err.Error())
					continue
				}

				pong := &model.Pong{IpAddr: ip, MACAddr: macAddr, RespTime: duration}

				s.pongs.Store(mynet.Ip2int(ip), pong)
				//s.logger.Debug().Msg(fmt.Sprintf("worker: %v,\tiface: %s,\tip: %s,\tmac: %s,\ttime: %s", num, iface.Name, ip, macAddr, duration))
				runtime.Gosched()
			}
		}(s.pingChan, i)
	}
}
