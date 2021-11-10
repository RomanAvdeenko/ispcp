package pinger

import (
	"database/sql"
	"fmt"

	"ispcp/internal/host"
	"ispcp/internal/model"
	"ispcp/internal/store"
	"ispcp/internal/store/file"
	"ispcp/internal/store/mysql"
	"os"
	"time"

	mynet "github.com/RomanAvdeenko/utils/net"

	"net"

	"github.com/j-keck/arping"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	f  *os.File
	db *sql.DB
	st store.Store
)

type Server struct {
	conifg   *Config
	store    store.Store
	logger   *zerolog.Logger
	host     *host.Host
	pingChan chan model.Ping
	pongs    *model.Pongs
	location *time.Location
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
func init() {
	arping.SetTimeout(10 * time.Millisecond)
	//arping.EnableVerboseLog()
}

func selectStoreType(cfg *Config, f *os.File, db *sql.DB) error {
	var err error
	if cfg.StoreType == "mysql" {
		// // Mysql store
		db, err = sql.Open("mysql", cfg.URI)
		if err != nil {
			fmt.Println("err: ", err)
			return err
		}

		if err = db.Ping(); err != nil {
			fmt.Println("err: ", err)
			return err
		}
		st = mysql.New(db)
	} else {
		// File store
		f, err = os.OpenFile("./store.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		st = file.New(f)
	}
	return nil
}

func Start(cfg *Config) error {
	if err := selectStoreType(cfg, f, db); err != nil {
		return err
	}
	defer db.Close()
	defer f.Close()

	s := newServer(cfg, st)

	refreshInterval := time.Duration(s.conifg.RestartInterval) * time.Second
	refreshTicker := time.NewTicker(refreshInterval)

	//s.logger.Info().Msg(fmt.Sprintf("Start pinger with %v threads, refresh interval: %s, store type: %s", s.conifg.ThreadsNumber, refreshInterval, s.conifg.StoreType))
	s.logger.Info().Msg(fmt.Sprintf("Start pinger with refresh interval: %s, store type: %s, logging level: %s", refreshInterval, s.conifg.StoreType, s.conifg.LoggingLevel))

	go func() {
		// Start working instantly
		s.addWork()
		s.startWorkers()

		for {
			select {
			case <-refreshTicker.C:
				// Check completion for previous work
				if len(s.pingChan) == 0 {
					s.logger.Info().Msg("Write to store")
					err := s.store.Store(s.pongs)
					if err != nil {
						s.logger.Error().Msg("Store error: " + err.Error())
						continue
					}
					s.pongs.Clear()
					s.addWork()
				} else {
					s.logger.Warn().Msg(fmt.Sprintf("Previous work has not been completed (%v IPs). Skip...", len(s.pingChan)))
				}
			}
		}
	}()

	quit := make(chan struct{})
	<-quit

	return nil
}

func (s *Server) configure() error {
	s.location, _ = time.LoadLocation("Europe/Kiev")
	s.configureLogger()

	s.host.SetExcludeIfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.SetExcludeNetworkIPs(s.conifg.ExcludeNetIPs)
	s.host.SetLogger(s.logger)
	s.host.Configure()

	return nil
}

func (s *Server) configureLogger() {
	*s.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	switch s.conifg.LoggingLevel {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "TRACE":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

}

// Adds work to ping all required host interfaces
func (s *Server) addWork() {
	go func(ch chan<- model.Ping) {
		defer s.logger.Debug().Msg("Adding work  done")

		s.logger.Debug().Msg("Starting to add work.")
		//Let's walk through the interfaces
		for _, iface := range s.host.ProcessedIfaces {
			ifaceAddrs, err := s.host.GetIfaceAddrs(iface)
			if err != nil {
				s.logger.Error().Msg("addWork(): " + err.Error())
				continue
			}
			for _, ifaceAddr := range ifaceAddrs {
				// Add job to workers
				func(iface net.Interface, addr string, ch chan<- model.Ping) {
					s.logger.Info().Msg(fmt.Sprintf("Processed interface: %v, processed network: %v", iface.Name, addr))
					ips, err := mynet.GetHostsIP(addr)
					if err != nil {
						s.logger.Error().Msg("mynet.GetHostsIP " + err.Error())
						return
					}
					for _, ip := range ips {
						s.logger.Debug().Msg("put: " + ip.String())
						ch <- model.Ping{IP: ip, Iface: iface}
					}
				}(iface, ifaceAddr, ch)
			}
		}
	}(s.pingChan)
}

func (s *Server) startWorkers() {
	// Bug!!! Only one
	//for i := 0; i < s.conifg.ThreadsNumber; i++ {
	// Start workers
	go func(ch <-chan model.Ping, num int) {
		defer s.logger.Error().Msg("Worker done")

		for ping := range ch {
			s.logger.Debug().Msg("receive: " + ping.IP.String())
			for c := 1; c < timesToRetry+1; c++ {
				s.logger.Trace().Msg(fmt.Sprintf("%s,\t%s.", ping.Iface.Name, ping.IP))
				MAC, duration, err := arping.PingOverIface(ping.IP, ping.Iface)
				if err != nil {
					if err != arping.ErrTimeout {
						// Try resend
						s.logger.Debug().Msg(fmt.Sprintf("Need to resend arp to %s. Try # %v of %v.", ping.IP, c, timesToRetry))
						time.Sleep(arpNanoSecDelay * time.Nanosecond)
						continue
					}
					s.logger.Trace().Msg(fmt.Sprintf("%s,\t%s: timeout.", ping.Iface.Name, ping.IP))
					pong := &model.Pong{IpAddr: ping.IP, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: false}
					s.pongs.Store(pong)
					break
				} else {
					s.logger.Trace().Msg(fmt.Sprintf("%s,\t%s: OK.", ping.Iface.Name, ping.IP))
					pong := &model.Pong{IpAddr: ping.IP, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: true}
					s.pongs.Store(pong)
					break
				}
				//s.logger.Debug().Msg(fmt.Sprintf("worker: %v,\tiface: %s,\tip: %s,\tmac: %s,\ttime: %s", num, ping.Iface.Name, ping.IP, macAddr, duration))
			}
		}
	}(s.pingChan, 0)
	//	}
}
