package postgres

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
)

type pgCfg struct {
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPass      string
	PostgresUriArgs   string
	PostgresDefaultDb string
	AnnotationFilter  string
	KeepSecretName    bool
}

var doOnce sync.Once
var config *pgCfg

func getConfig() *pgCfg {
	doOnce.Do(func() {
		config = &pgCfg{}
		config.PostgresHost = MustGetEnv("POSTGRES_HOST")
		config.PostgresHost = MustGetEnv("POSTGRES_PORT")
		config.PostgresUser = url.PathEscape(MustGetEnv("POSTGRES_USER"))
		config.PostgresPass = url.PathEscape(MustGetEnv("POSTGRES_PASS"))
		config.PostgresUriArgs = MustGetEnv("POSTGRES_URI_ARGS")
		config.PostgresDefaultDb = GetEnv("POSTGRES_DEFAULT_DATABASE")
		config.AnnotationFilter = GetEnv("POSTGRES_INSTANCE")
		if value, err := strconv.ParseBool(GetEnv("KEEP_SECRET_NAME")); err == nil {
			config.KeepSecretName = value
		}
	})
	return config
}

func MustGetEnv(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		log.Fatalf("environment variable %s is missing", name)
	}
	return value
}

func GetEnv(name string) string {
	return os.Getenv(name)
}