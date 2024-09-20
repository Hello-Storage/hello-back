package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/joho/godotenv"
)

type EnvVar struct {
	// App env
	AppPort string
	AppEnv  string
	// token env
	TokenSymmetricKey    string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	// Postgres env
	DBHost     string
	DBName     string
	DBUser     string
	DBPassword string
	DBPort     string
	// Github OAuth credential
	//GithubClientID     string
	//GithubClientSecret string
	// storage keys
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageEndpoint  string
	StorageRegion    string
	// Encryption
	EncryptionKey string
	// MailGun
	MailGunApiKey string
	// Arweave
	NodejsServerEndpoint string
}

var env EnvVar

func LoadEnv() (err error) {
	// skip load env when docker
	if os.Getenv("APP_PORT") == "" {
		err = godotenv.Load(".env")
		if err != nil {
			return err
		}
	}

	atd, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		return err
	}

	rtd, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		return err
	}

	env = EnvVar{
		// App env
		AppPort: os.Getenv("APP_PORT"),
		AppEnv:  os.Getenv("APP_ENV"),
		// token env
		TokenSymmetricKey:    os.Getenv("TOKEN_SYMMETRIC_KEY"),
		AccessTokenDuration:  atd,
		RefreshTokenDuration: rtd,
		// Postgres
		DBHost:     os.Getenv("POSTGRES_HOST"),
		DBName:     os.Getenv("POSTGRES_DB"),
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBPort:     os.Getenv("POSTGRES_PORT"),
		// Github OAuth credentail
		//GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		//GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		//Storage keys
		StorageAccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
		StorageSecretKey: os.Getenv("STORAGE_SECRET_KEY"),
		StorageBucket:    os.Getenv("STORAGE_BUCKET"),
		StorageEndpoint:  os.Getenv("STORAGE_ENDPOINT"),
		StorageRegion:    os.Getenv("STORAGE_REGION"),
		// Encryption
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		// MailGun
		MailGunApiKey: os.Getenv("MAILGUN_API"),
		// Arweave
		NodejsServerEndpoint: os.Getenv("NODEJS_SERVER_ENDPOINT"),
	}

	values := reflect.ValueOf(env)
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		if values.Field(i).String() == "" {
			return fmt.Errorf("config: %s is missing", types.Field(i).Name)
		}
	}

	return
}

func Env() EnvVar {
	return env
}
