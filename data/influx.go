package data

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func StartWriteInfluxHTTPV2(cfg Config, f func(logtext interface{}), InputString chan string) {
	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	writeAPI := client.WriteAPIBlocking(cfg.InfluxDBORG, cfg.InfluxDBBucket)

	defer client.Close()
	for str := range InputString {

		err := writeAPI.WriteRecord(context.Background(), str)
		if err != nil {
			f(err)
		}
	}
}

func StartWriteInfluxHTTPV1(cfg Config, f func(logtext interface{}), InputString chan string) {

	request := cfg.InfluxDBURL

	if !strings.HasSuffix(request, "/") {
		request += "/"
	}

	request += "write?db=" + cfg.InfluxDBBucket

	for str := range InputString {

		resp, err := http.NewRequest("POST", request, nil)
		if err != nil {
			f(err)
		}

		resp.Header.Add("User-Agent", "go-summazier_jmeter")
		resp.Body = io.NopCloser(strings.NewReader(str + " " + fmt.Sprintf("%d", time.Now().UnixNano())))

		cli := &http.Client{}
		rsp, err := cli.Do(resp)
		if err != nil {
			f(err)
		}

		if rsp.StatusCode <= 200 || rsp.StatusCode >= 299 {
			f("Write error " + strconv.Itoa(rsp.StatusCode) + " " + request + str)
		}

		cli.CloseIdleConnections()
	}
}

func StartWriteInfluxUDPV1(cfg Config, f func(logtext interface{}), InputString chan string) {

	conn, err := NewUDPClient(cfg)
	if err != nil {
		f(err)
	}

	defer conn.Close()

	for str := range InputString {
		// В каком формате UDP?
		_, err := conn.Write([]byte(str))
		if err != nil {
			f(err)
		}
		f(str)
	}
}

func NewUDPClient(cfg Config) (*net.UDPConn, error) {
	var udpAddr *net.UDPAddr
	var url string
	if strings.HasPrefix(cfg.InfluxDBURL, "http://") {
		url = cfg.InfluxDBURL[len("http://"):]
	} else {
		url = cfg.InfluxDBURL
	}

	udpAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}
