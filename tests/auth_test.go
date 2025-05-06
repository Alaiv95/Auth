package tests

import (
	"Auth/tests/suite"
	authv1 "github.com/Alaiv95/Protos/gen/go/auth"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

const (
	emptyAppID     = 0
	appID          = 1
	appSecret      = "test-secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_Success(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := fakePassword()

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)

	respLogin, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})

	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, appID, int(claims["app_id"].(float64)))
	assert.Equal(t, email, claims["email"])
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTtl).Unix(), claims["exp"].(float64), 1)
}

func TestRegisterLogin_Login_AppNotFound(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := fakePassword()

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)

	resp, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    emptyAppID,
	})

	require.Error(t, err)
	assert.Empty(t, resp)
}

func TestRegister_Register_UserExists(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := fakePassword()

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	_, err = st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "User already exists"))
}

func fakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
