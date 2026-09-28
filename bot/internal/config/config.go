// package config - парсинг конфиг-файлов и
// раздача их содержимого в другие скрипты
package config

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MaxToken          string
	BaseURL           string
	MiniAppWebApp     string
	MiniAppStartParam string
}

// LoadConfig - читает .env, если он есть, и забирает переменные окружения.
// В контейнере файла нет: токен и адрес приходят из environment при старте.
func LoadConfig() *Config {
	for _, path := range []string{"../.env", "../../../.env", ".env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	token := strings.TrimSpace(os.Getenv("MAX_TOKEN"))
	if token == "" {
		log.Fatal("MAX_TOKEN отсутствует в окружении")
	}

	baseURL := normalizeBaseURL(os.Getenv("BASE_URL"))
	if baseURL == "" {
		log.Fatal("BASE_URL отсутствует в .env файле")
	}

	return &Config{
		MaxToken:          token,
		BaseURL:           baseURL,
		MiniAppWebApp:     strings.TrimSpace(os.Getenv("MAX_MINIAPP_WEB_APP")),
		MiniAppStartParam: strings.TrimSpace(os.Getenv("MAX_MINIAPP_START_PARAM")),
	}
}

func normalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return strings.TrimRight(raw, "/")
	}
	if host, err := url.PathUnescape(u.Host); err == nil {
		u.Host = host
	}
	if u.Host == "api.арктикпаспорт.рф" {
		u.Host = "api.xn--80aaxhbzfbdhihi.xn--p1ai"
	}
	return strings.TrimRight(u.String(), "/")
}
