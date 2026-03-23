package http

import (
	"fmt"
	"net"
	"strconv"
)

func ReadLocalIP() (string, error) {
	addr, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, address := range addr {
		if in, ok := address.(*net.IPNet); ok && !in.IP.IsLoopback() {
			if in.IP.To4() != nil {
				return in.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("can't find the address")
}

func GetServerAddress(port int) string {
	return LocalIPWithOutError() + ":" + strconv.Itoa(port)
}

func LocalIPWithOutError() string {
	ip, _ := ReadLocalIP()
	return ip
}

func ResolveHost(host string) (ip string, err error) {
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return "", err
	}
	return address.String(), err
}

func SplitHostPort(address string) (host string, port int, err error) {
	var strPort string

	if host, strPort, err = net.SplitHostPort(address); err != nil {
		return "", -1, err
	}

	if port, err = strconv.Atoi(strPort); err != nil {
		return "", -1, err
	}

	if host == "" {
		return "", port, fmt.Errorf("host is empty, please contact DBA")
	}

	return host, port, nil
}
