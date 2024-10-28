package data

import (
	"context"
	"fmt"
	"io/ioutil"
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
	for {
		select {
		case str := <-InputString:
			err := writeAPI.WriteRecord(context.Background(), str)
			if err != nil {
				f(err)
			}
		default:
		}
	}
}

func StartWriteInfluxHTTPV1(cfg Config, f func(logtext interface{}), InputString chan string) {
	heartbeat := time.Tick(10 * time.Second)

	for {
		select {
		case <-heartbeat:
			f("OK")

		case str := <-InputString:
			request := cfg.InfluxDBURL

			resp, err := http.NewRequest("POST", request, nil)
			if err != nil {
				f(err)
			}

			resp.Header.Add("User-Agent", "go-summazier_jmeter")
			resp.Body = ioutil.NopCloser(strings.NewReader(str + " " + fmt.Sprintf("%d", time.Now().UnixNano())))

			cli := &http.Client{}
			rsp, err := cli.Do(resp)

			if err != nil {
				f(err)
			}

			if rsp.StatusCode <= 200 || rsp.StatusCode >= 299 {
				f("Write error " + strconv.Itoa(rsp.StatusCode) + " " + request + str)
			}

			defer cli.CloseIdleConnections()
		default:
		}
	}
}

func StartWriteInfluxUDPV1(cfg Config, f func(logtext interface{}), InputString chan string) {
	heartbeat := time.Tick(10 * time.Second)

	conn, err := NewUDPClient(cfg)
	if err != nil {
		f(err)
	}

	defer conn.Close()

	for {
		select {
		case <-heartbeat:
			f("OK")

		case str := <-InputString:
			// В каком формате UDP?
			_, err := conn.Write([]byte(str))
			if err != nil {
				f(err)
			}
			f(str)
		}
	}
}

func NewUDPClient(cfg Config) (*net.UDPConn, error) {
	var udpAddr *net.UDPAddr
	udpAddr, err := net.ResolveUDPAddr("udp", cfg.InfluxDBURL)

	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}
