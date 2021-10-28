package pinger

import (
	"ispcp/internal/host"
	"ispcp/internal/model"
	"log"
	"net"
	"os"

	"github.com/j-keck/arping"
)

type PingerServer struct {
	conifg *Config
	logger *log.Logger
	host   *host.Host
}

func newServer(cfg *Config) *PingerServer {
	s := PingerServer{
		conifg: cfg,
		logger: log.New(os.Stdout, "", log.Lshortfile),
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
	//
	s.host.SetExcludeInterfaceNames(s.conifg.ExcludeIfaceNames)
	s.host.Configure()
	return nil
}
func (s *PingerServer) start() error {
	s.logger.Println("Start pinger...")
	return nil
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
