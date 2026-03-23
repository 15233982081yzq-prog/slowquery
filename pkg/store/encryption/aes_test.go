package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAes(t *testing.T) {
	origin := "1234567654321"
	errKey := "^^"
	key := "buhxw6cZ4*YFVkpv"
	_, err := AesEncryptCBC([]byte(origin), []byte(errKey))
	assert.NotNil(t, err)
	en, err := AesEncryptCBC([]byte(origin), []byte(key))
	assert.Nil(t, err)
	or, err := AesDecryptCBC(en, []byte(key))
	assert.Nil(t, err)
	assert.Equal(t, string(or), origin)
	assert.Equal(t, or, []byte(origin))
	println("---------------------------------")
	println(origin)
	println(string(en))
	println(string(or))
}
