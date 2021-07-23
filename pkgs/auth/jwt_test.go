package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJwt(t *testing.T) {
	s, err := GenJwtToken("52341i7367", "longalong", "longalong@longalong.cn")
	fmt.Printf("err : %s\n", err)

	assert.Nil(t, err)
	fmt.Printf("token: %s\n", s)

	e, err := JwtTokenValidate(s)
	fmt.Printf("err : %s\n", err)

	assert.Nil(t, err)
	fmt.Printf("email : %s\n", e)
}
