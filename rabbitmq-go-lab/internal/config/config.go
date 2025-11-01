package config

import (
	"os"
	"strconv"
)

type C struct {
	URL         string
	User        string
	Pass        string
	VHost       string
	Host        string
	Port        string
	Exchange    string
	RetryExchange string
	DLX         string
	Queue       string
	RetryQueue  string
	DLQ         string
	RetryTTLms  int
	MaxRetries  int
}

func Env(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}

func MustInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil { return n }
	}
	return def
}

func Load() C {
	c := C{
		User: Env("RABBITMQ_USER", "guest"),
		Pass: Env("RABBITMQ_PASS", "guest"),
		Host: Env("RABBITMQ_HOST", "localhost"),
		Port: Env("RABBITMQ_PORT", "5672"),
		VHost: Env("RABBITMQ_VHOST", "/"),
		Exchange: Env("APP_EXCHANGE", "app.events"),
		RetryExchange: Env("APP_RETRY_EXCHANGE", "app.events.retry"),
		DLX: Env("APP_DLX", "app.events.dlx"),
		Queue: Env("APP_QUEUE", "app.events.main"),
		RetryQueue: Env("APP_RETRY_QUEUE", "app.events.retry"),
		DLQ: Env("APP_DLQ", "app.events.dlq"),
		RetryTTLms: MustInt("RETRY_TTL_MS", 10000),
		MaxRetries: MustInt("MAX_RETRIES", 3),
	}
	c.URL = "amqp://" + c.User + ":" + c.Pass + "@" + c.Host + ":" + c.Port + c.VHost
	return c
}
