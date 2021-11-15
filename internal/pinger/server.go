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
	logger   *zerolog.Logger
	host     *host.Host
	pongs    *model.Pongs
	location *time.Location
	run      bool
}

func newServer(cfg *Config, store store.Store) *Server {
	s := Server{
		conifg: cfg,
		store:  store,
		//logger: &zerolog.New(os.Stdout),
		logger: &zerolog.Logger{},
		host:   host.NewHost(),
		pongs:  model.NewPongs(),
	}
	s.configure()
	return &s
}
func init() {
	arping.SetTimeout(30 * time.Millisecond)
	//arping.EnableVerboseLog()
}

func Stop() {
	fmt.Println("Terminate...")
	if db != nil {
		db.Close()
	}
	if f != nil {
		f.Close()
	}
}

func Start(cfg *Config) error {
	defer Stop()

	if err := selectStoreType(cfg, f, db); err != nil {
		return err
	}
	s := newServer(cfg, st)

	refreshInterval := time.Duration(s.conifg.RestartInterval) * time.Second
	refreshTicker := time.NewTicker(refreshInterval)

	s.logger.Info().Msg(fmt.Sprintf("Start pinger with refresh interval: %s, store type: %s, logging level: %s", refreshInterval, s.conifg.StoreType, s.conifg.LoggingLevel))
	// Start working instantly
	go s.Do()
	for range refreshTicker.C {
		if !s.run {
			go s.Do()
		} else {
			s.logger.Warn().Msg("Can't start, previouswork isn't finished!")
		}
	}
	return nil
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

// Adds work to ipl required host interfaces
func (s *Server) Do() {
	defer func() {
		s.pongs.Clear()
		s.run = false
	}()

	s.logger.Debug().Msg("Starting to add work.")
	s.run = true
	//Let's walk through the interfaces
	for _, iface := range s.host.ProcessedIfaces {
		ifaceAddrs, err := s.host.GetIfaceAddrs(iface)
		if err != nil {
			s.logger.Error().Msg("addWork(): " + err.Error())
			continue
		}
		for _, addr := range ifaceAddrs {
			s.logger.Info().Msg(fmt.Sprintf("Processed interface: %v, processed network: %v", iface.Name, addr))
			ips, err := mynet.GetHostsIP(addr)
			if err != nil {
				s.logger.Error().Msg("mynet.GetHostsIP " + err.Error())
				return
			}
			for _, ip := range ips {
				// Resend if error
				for c := 1; c < timesToRetry+1; c++ {
					MAC, duration, err := arping.PingOverIface(ip, iface)
					s.logger.Trace().Msg(fmt.Sprintf("%v,\t%v.", iface, ip))

					s.logger.Trace().Msg(fmt.Sprintf("%v,\t%v\t%v.", MAC, duration, err))
					if err != nil {
						if err != arping.ErrTimeout {
							// Try to resend
							s.logger.Debug().Msg(fmt.Sprintf("Resend arp to %s, %v of %v.", ip, c, timesToRetry))
							time.Sleep(arping.ArpDelay)
							continue
						}
						s.logger.Trace().Msg(iface.Name + ip.String() + " timeout.")
						pong := &model.Pong{IpAddr: ip, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: false}
						s.pongs.Store(pong)
						break
					} else {
						s.logger.Trace().Msg(fmt.Sprintf("%s,\t%s: OK.", iface.Name, ip))
						pong := &model.Pong{IpAddr: ip, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: true}
						s.pongs.Store(pong)
						break
					}
				}
				time.Sleep(arping.ArpDelay)
			}
		}
	}
	s.logger.Debug().Msg("Work  done.")

	s.logger.Info().Msg("Write to the store.")
	err := s.store.Store(s.pongs)
	if err != nil {
		s.logger.Error().Msg("Store error: " + err.Error())
		return
	}
}
