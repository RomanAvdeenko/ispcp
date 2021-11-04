package pinger

import (
	"fmt"
	mynet "github.com/RomanAvdeenko/utils/net"
	"ispcp/internal/host"
	"ispcp/internal/model"
	"ispcp/internal/store"
	"ispcp/internal/store/file"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"net"

	"github.com/j-keck/arping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var macRegexp = regexp.MustCompile(`[a-fA-F0-9:]{17}|[a-fA-F0-9]{12}`)

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
	f, err := os.OpenFile("./store.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	store := file.New(f)
	s := newServer(cfg, store)

	refreshInterval := time.Duration(s.conifg.RestartInterval) * time.Second
	refreshTicker := time.NewTicker(refreshInterval)

	//
	arping.SetTimeout(10 * time.Second)
	//

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
					s.store.Store(s.pongs)
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
	s.configureLogger()

	s.host.SetExcludeIfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.SetExcludeNetworkIPs(s.conifg.ExcludeNetIPs)
	s.host.SetLogger(s.logger)
	s.host.Configure()

	return nil
}

func (s *Server) configureLogger() {
	*s.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
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

				//macAddr, duration, err := arping.PingOverIface(ping.IP, ping.Iface)

				ip := ping.IP.String()

				args := []string{"-I", ping.Iface.Name, ip, "-c1"}
				//args := []string{"-i", ping.Iface.Name, ip}

				cmd := "/usr/bin/arping"
				//cmd := "/home/rav/go/bin/arpin"

				//s.logger.Printf("%s %s", cmd, args)

				out, err := exec.Command(cmd, args...).CombinedOutput()
				//s.logger.Error().Msg(err.Error() + " " + string(out))

				//time.Sleep(20 * time.Millisecond)

				if err != nil {
					if err != arping.ErrTimeout && string(out) != "timeout\n" {
						//s.logger.Printf("%s,\t%s,\t%s,\t\t%s", ping.Iface.Name, ping.IP, err, out)
						//s.logger.Error().Msg(string(out))
					}
					continue
				}
				//

				MAC, _ := net.ParseMAC(macRegexp.FindString(string(out)))
				//
				//pong := &model.Pong{IpAddr: ping.IP, MACAddr: macAddr, Time: time.Now(), Duration: duration, Alive: true}
				pong := &model.Pong{IpAddr: ping.IP, Time: time.Now(), Alive: true, MACAddr: MAC}
				s.pongs.Store(pong)
				//s.logger.Debug().Msg(fmt.Sprintf("worker: %v,\tiface: %s,\tip: %s,\tmac: %s,\ttime: %s", num, ping.Iface.Name, ping.IP, macAddr, duration))
				runtime.Gosched()
			}
		}(s.pingChan, i)
	}
}
