package http

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveHost(t *testing.T) {
	add, err := ResolveHost("asdadadadadad")
	assert.Errorf(t, err, "ResolveHost error")
	assert.Empty(t, add)

	add, err = ResolveHost("space.test.shopee.io")
	assert.NoError(t, err)
	assert.NotEmpty(t, add)
}

func TestReadLocalIP(t *testing.T) {
	ip, err := ReadLocalIP()
	assert.NoError(t, err)
	fmt.Printf("local ip ip:%s \n", ip)
}

func TestGetServerAddress(t *testing.T) {
	address := GetServerAddress(3223)
	assert.NotNil(t, address)
	fmt.Printf("server_address:%s \n", address)
}

func TestSplitHostPort(t *testing.T) {
	okAddress := GetServerAddress(3223)
	ip, port, err := SplitHostPort(okAddress)
	assert.NoError(t, err)
	fmt.Printf("SplitHostPort(%s) ,ip:%s,port:%d \n", okAddress, ip, port)

	noPort := "10.2.111.2.ku22"
	ip, port, err = SplitHostPort(noPort)
	assert.Errorf(t, err, fmt.Sprintf("adrress:%s SplitHostPort error:%s \n", noPort, err.Error()))

	failPort := "1.1.1.1:euwt"
	ip, port, err = SplitHostPort(failPort)
	assert.Error(t, err)

	onlyPort := ":12345"
	ip, port, err = SplitHostPort(onlyPort)
	assert.Error(t, err)
}

func TestLocalIPWithOutError(t *testing.T) {
	ip := LocalIPWithOutError()
	assert.NotNil(t, ip)
}
