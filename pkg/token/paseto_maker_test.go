package token

import (
	"testing"
	"time"

	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(rnd.GenerateRandomString(32))
	require.NoError(t, err)

	user_uid := rnd.GenerateRandomString(6)
	user_name := rnd.GenerateRandomString(6)
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(1, user_uid, user_name, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.NotZero(t, payload.TokenID)
	require.Equal(t, user_name, payload.UserName)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(rnd.GenerateRandomString(32))
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(
		1,
		rnd.GenerateRandomString(6),
		rnd.GenerateRandomString(6),
		-time.Minute,
	)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
