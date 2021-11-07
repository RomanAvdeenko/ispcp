package pinger

import (
	"fmt"
	"strconv"

	"ispcp/internal/host"
	"ispcp/internal/model"
	"ispcp/internal/store"
	"ispcp/internal/store/file"
	"os"
	"time"

	mynet "github.com/RomanAvdeenko/utils/net"

	"net"

	"github.com/j-keck/arping"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
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

func Start(cfg *Config) error {
	//arping.SetTimeout(100 * time.Millisecond)
	//arping.EnableVerboseLog()

	// File store
	f, err := os.OpenFile("./store.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	store := file.New(f)

	// // Mysql store
	// db, err := sql.Open("mysql", cfg.URI)
	// if err != nil {
	// 	fmt.Println("err: ", err)
	// 	return err
	// }
	// defer db.Close()

	// if err := db.Ping(); err != nil {
	// 	fmt.Println("err: ", err)
	// 	return err
	// }

	// store := mysql.New(db)
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
				// Check completion for previous work
				if len(s.pingChan) == 0 {
					err := s.store.Store(s.pongs)
					if err != nil {
						s.logger.Error().Msg("Store error: " + err.Error())
						continue
					}
					s.pongs.Clear()
					s.addWork()
					s.startWorkers()
				} else {
					s.logger.Info().Msg("Check completion of previous work. Previous work has not been completed. Skip.")
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
	s.logger.Level(zerolog.DebugLevel)
}

// Adds work to ping all required host interfaces
func (s *Server) addWork() error {
	s.logger.Debug().Msg("Starting to add work.")
	//Let's walk through the interfaces
	for _, iface := range s.host.ProcessedIfaces {
		ifaceAddrs, err := s.host.GetIfaceAddrs(iface)
		if err != nil {
			return err
		}
		for _, ifaceAddr := range ifaceAddrs {
			// Add job to workers
			go func(iface net.Interface, addr string, pingChan chan<- model.Ping) {
				s.logger.Info().Msg(fmt.Sprintf("Processed interface: %v, processed network: %v", iface.Name, addr))
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
		go func(pingChan chan model.Ping, num int) {
			defer time.Sleep(arpNanoSecDelay * time.Nanosecond)

			for ping := range pingChan {
				var alive bool
				for c := 0; c < timesToRetry; c++ {
					MAC, duration, err := arping.PingOverIface(ping.IP, ping.Iface)
					if err != nil {
						if err != arping.ErrTimeout {
							// Try resend
							s.logger.Info().Msg(ping.IP.String() + " need to send again " + strconv.Itoa(c))
							time.Sleep(arpNanoSecDelay * time.Nanosecond)
							continue
						}
						break
					} else {
						alive = true
						pong := &model.Pong{IpAddr: ping.IP, MACAddr: MAC, Time: time.Now().In(s.location), Duration: duration, Alive: alive}
						s.pongs.Store(pong)
						break
					}
					//s.logger.Printf("%s,\t%s,\t%s,\t\t%s", ping.Iface.Name, ping.IP, "OK", "")
					//s.logger.Debug().Msg(fmt.Sprintf("worker: %v,\tiface: %s,\tip: %s,\tmac: %s,\ttime: %s", num, ping.Iface.Name, ping.IP, macAddr, duration))
				}
			}
		}(s.pingChan, i)
	}
}
