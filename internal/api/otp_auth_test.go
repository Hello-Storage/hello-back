package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

func TestOTP(t *testing.T) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "hello.app",
		AccountName: "alice@example.com",
	})
	require.NoError(t, err)

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	fmt.Println("code:", code)
	require.NoError(t, err)

	result := totp.Validate(code, key.Secret())
	fmt.Println("secret:", key.Secret())
	require.Equal(t, result, true)
}
