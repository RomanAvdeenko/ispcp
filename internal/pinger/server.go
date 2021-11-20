package pinger

import (
	"database/sql"
	"fmt"
	"syscall"

	"ispcp/internal/host"
	"ispcp/internal/model"
	"ispcp/internal/store"
	"ispcp/internal/store/file"
	"ispcp/internal/store/mysql"
	"os"
	"time"

	mynet "github.com/RomanAvdeenko/utils/net"

	//"github.com/j-keck/arping"
	"ispcp/internal/arping"

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
	host     *host.Host
	pongs    *model.Pongs
	location *time.Location
	run      int // 0 - stop, over - number of start attempts
}

func newServer(cfg *Config, store store.Store) *Server {
	s := Server{
		conifg: cfg,
		store:  store,
		host:   host.NewHost(),
		pongs:  model.NewPongs(),
	}
	s.configure()
	return &s
}
func init() {
	//arping.EnableVerboseLog()
}

func Clean() {
	if db != nil {
		db.Close()
	}
	if f != nil {
		f.Close()
	}
}

func Start(cfg *Config) error {
	var err error
	responseWaitTime := time.Millisecond * time.Duration(cfg.ResponseWaitTime)
	arping.SetTimeout(responseWaitTime)

	if err := selectStoreType(cfg, f, db); err != nil {
		return err
	}

	s := newServer(cfg, st)
	s.location, err = time.LoadLocation(cfg.Location)
	if err != nil {
		log.Error().Msg("Couldn't set locale from config. Installed " + locationDefault)
		s.location, _ = time.LoadLocation(locationDefault)
	}

	refreshInterval := time.Duration(s.conifg.RestartInterval) * time.Second
	refreshTicker := time.NewTicker(refreshInterval)

	log.Info().Msg(fmt.Sprintf("-->Start pinger with refresh interval: %s, response wait time: %s, store type: %s, logging level: %s",
		refreshInterval, responseWaitTime, s.conifg.StoreType, s.conifg.LoggingLevel))
	// Start working instantly
	go s.Do()
	for range refreshTicker.C {
		if s.run == 0 {
			go s.Do()
		} else {
			if s.run < timesToMainLoopRetry {
				s.run++
				log.Warn().Msg("Can't start, previouswork isn't finished!")
			} else {
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		}
	}
	return nil
}

func selectStoreType(cfg *Config, fi *os.File, db *sql.DB) error {
	var err error
	if cfg.StoreType == "mysql" {
		// // Mysql store
		db, err = sql.Open("mysql", cfg.URI)
		if err != nil {
			return err
		}
		if err = db.Ping(); err != nil {
			return err
		}
		st = mysql.New(db)
	} else {
		// File store
		fi, err = os.OpenFile("./store.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		st = file.New(fi)
	}
	return nil
}

func (s *Server) configure() error {
	s.configureLogger()
	s.host.SetExcludeIfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.SetExcludeNetworkIPs(s.conifg.ExcludeNetIPs)
	s.host.Configure()
	return nil
}

func (s *Server) configureLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
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

// Adds work to ipl required host interfaces
func (s *Server) Do() {
	defer func() {
		s.pongs.Clear()
		s.run = 0
	}()

	log.Debug().Msg("->Starting to add work.")
	s.run = 1
	// Сonfiguration may change
	//...
	s.configure()

	//Let's walk through the interfaces
	for _, iface := range s.host.ProcessedIfaces {
		ifaceAddrs, err := s.host.GetIfaceAddrs(iface)
		if err != nil {
			log.Error().Msg("addWork(): " + err.Error())
			continue
		}
		for _, addr := range ifaceAddrs {
			log.Info().Msg(fmt.Sprintf("Processed interface: %v, processed network: %v", iface.Name, addr))
			ips, err := mynet.GetHostsIP(addr)
			if err != nil {
				log.Error().Msg("mynet.GetHostsIP " + err.Error())
				return
			}
			for _, ip := range ips {
				// Resend if error
				for c := 1; c < timesToArpIPRetry+1; c++ {
					MAC, duration, err := arping.PingOverIface(ip, iface)
					log.Trace().Msg(fmt.Sprintf("%v,\t%v.", iface, ip))

					log.Trace().Msg(fmt.Sprintf("%v,\t%v\t%v.", MAC, duration, err))
					if err != nil {
						if err != arping.ErrTimeout {
							// Try to resend
							log.Debug().Msg(fmt.Sprintf("Resend arp to %s, %v of %v.", ip, c, timesToArpIPRetry))
							continue
						}
						log.Trace().Msg(iface.Name + ip.String() + " timeout.")
						pong := &model.Pong{IpAddr: ip, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: false}
						s.pongs.Store(pong)
						break
					} else {
						log.Trace().Msg(fmt.Sprintf("%s,\t%s: OK.", iface.Name, ip))
						pong := &model.Pong{IpAddr: ip, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: true}
						s.pongs.Store(pong)
						break
					}
				}
			}
		}
	}
	log.Info().Msg("Writing to the store.")
	err := s.store.Store(s.pongs)
	if err != nil {
		log.Error().Msg("Store error: " + err.Error())
		return
	}
	log.Debug().Msg("->Work  done.")
}
