package pinger

import (
	"fmt"
	"ispcp/internal/host"
	"ispcp/internal/model"
	"os"
	"time"

	"net"

	"github.com/j-keck/arping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PingerServer struct {
	conifg *Config
	logger *zerolog.Logger
	host   *host.Host
}

func newServer(cfg *Config) *PingerServer {
	s := PingerServer{
		conifg: cfg,
		//logger: zerolog.New(os.Stdout),
		logger: &zerolog.Logger{},
		host:   host.NewHost(),
	}
	s.configure()
	return &s
}

func Start(cfg *Config) error {
	s := newServer(cfg)
	//
	return s.start()
}

func (s *PingerServer) configure() error {
	s.configureLogger()

	s.host.SetExcludeIfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.SetExcludeNetworkIPs(s.conifg.ExcludeNetIPs)
	s.host.SetLogger(s.logger)
	s.host.Configure()

	return nil
}

func (s *PingerServer) start() error {
	s.logger.Info().Msg("Start pinger...")
	ips, _ := s.host.GetIfaceAddresses(s.host.ProcessedIfaces[0])
	fmt.Println(ips)
	s.logger.Info().Msg("Stop pinger...")
	return nil
}

func (s *PingerServer) configureLogger() {
	*s.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli})
}

func (s *PingerServer) ping(pingChan <-chan string, pongChan chan<- model.Pong, ifaceName string) {
	for ip := range pingChan {
		ipAddr := net.ParseIP(ip)
		macAddr, duration, err := arping.PingOverIfaceByName(ipAddr, ifaceName)

		var alive bool
		if err != nil {
			//log.Println(ip, err)
			alive = false
		} else {
			alive = true
		}
		pongChan <- model.Pong{IpAddr: ipAddr, Alive: alive, RespTime: duration, MacAddr: macAddr}
	}
}

func receivePong(pongNum int, pongChan <-chan model.Pong, doneChan chan<- []model.Pong) {
	var alives []model.Pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func main1() {
	/*	//arping.EnableVerboseLog()
		hosts, _ := mynet.GetHosts("192.168.1.1/24")

		pingChan := make(chan string, 4)
		pongChan := make(chan model.Pong, len(hosts))
		doneChan := make(chan []model.Pong)

		// Start workers
		for i := 0; i < s.ThreadsNumber; i++ {
			go ping(pingChan, pongChan, "wlo1")
		}

		// Set job
		go func() {
			for _, ip := range hosts {
				pingChan <- ip
				//fmt.Println("sent: " + ip)
			}
			close(pingChan)
		}()

		// Start Receiver
		go receivePong(len(hosts), pongChan, doneChan)

		// Get results
		alives := <-doneChan
		for _, alive := range alives {
			fmt.Println(alive)
		}
	*/
}
