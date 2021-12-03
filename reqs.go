package main

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/logrusorgru/aurora/v3"
)

var (
	statusCode []string
	recv       []time.Time
	emailGot   string
	proxys     []string
	proxyList  []*url.URL
)

func testSpeed(conn *tls.Conn, number int) {
	e := make([]byte, 1000)
	conn.Read(e)
	recv = append(recv, time.Now())
	statusCode = append(statusCode, string(e[9:12]))
	if string(e[9:12]) == `200` {
		emailGot = emails[number]
	}
}

// - gc Sockets! -

var accountNum int

func gcReq() ([]time.Time, []time.Time, []string) {
	sendTime := make([]time.Time, 0)
	var js = []byte(`{"profileName":"` + name + `"}`)
	length := strconv.Itoa(len(string(js)))
	var conns []*tls.Conn
	payload := make([]string, 0)

	preSleep(dropTime)

	for i := 0; i < len(approvedBearer); i++ {
		conn, _ := tls.Dial("tcp", "api.minecraftservices.com"+":443", nil)
		conns = append(conns, conn)
	}

	for _, bearer := range approvedBearer {
		payload = append(payload, fmt.Sprintf("POST /minecraft/profile HTTP/1.1\r\nHost: api.minecraftservices.com\r\nConnection: open\r\nContent-Length:%s\r\nContent-Type: application/json\r\nAccept: application/json\r\nAuthorization: Bearer %s\r\n\r\n"+string(js)+"\r\n", length, bearer))
	}
	sleep(dropTime)

	for e, conn := range conns {
		for i := 0; float64(i) < config[`GcReq`].(float64); {
			fmt.Fprintln(conn, payload[e])
			sendTime = append(sendTime, time.Now())
			go testSpeed(conn, e)
			i++
			time.Sleep(time.Duration(config["SpreadPerReq"].(float64)) * time.Microsecond)
		}
		e++
	}

	time.Sleep(2 * time.Second)

	return sendTime, recv, statusCode
}

func preSleep(dropTime int64) {
	fmt.Print("\n")
	dropStamp := time.Unix(dropTime, 0)
	delDroptime := dropStamp.Add(-time.Second * 5)

	for {
		fmt.Printf(aurora.Sprintf(aurora.White(aurora.Bold(name+" | Dropping in %v    \r")), aurora.Bold(aurora.Red(time.Until(delDroptime).Round(time.Second).Seconds()))))
		time.Sleep(time.Second * 1)
		if time.Until(dropStamp) <= 5*time.Second {
			break
		}
	}
}

func sleep(dropTime int64) {
	dropStamp := time.Unix(dropTime, 0)

	fmt.Print(aurora.White(aurora.Bold("\n\nPreparing to Snipe...\n\n")))

	// theres 2 options because im testing them.. feel free to swap them urself
	switch config[`UseNanoSleep`] {
	case true:
		time.Sleep(time.Until(dropStamp.Add(time.Millisecond * time.Duration(0-delay)).Add(time.Duration(-float64(time.Since(time.Now()).Nanoseconds())/1000000.0) * time.Millisecond)))
	case false:
		time.Sleep(time.Until(dropStamp.Add(time.Millisecond * time.Duration(0-delay))))
	}
}
