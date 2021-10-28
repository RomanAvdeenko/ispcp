package pinger

import (
	"fmt"
	"ispcp/internal/model"
	"net"

	mynet "github.com/RomanAvdeenko/utils/net"
	"github.com/j-keck/arping"
)

const concurrentMax = 4

func ping(pingChan <-chan string, pongChan chan<- model.Pong, ifaceName string) {
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
	//arping.EnableVerboseLog()
	hosts, _ := mynet.GetHosts("46.162.42.0/24")

	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan model.Pong, len(hosts))
	doneChan := make(chan []model.Pong)

	// Start workers
	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan, "eth3.72")
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
}
