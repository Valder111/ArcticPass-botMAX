// package max - client - обёртка над maxbot.New() и HTTP-клиент к API MAX
package max

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const defaultHTTPTimeout = 60 * time.Second

//go:embed certs/russian_trusted_ca.pem
var embeddedRussianTrustedCA []byte

// NewBotClient - собирает клиент MAX API с TLS и токеном бота
func NewBotClient(token string) (*maxbot.Api, error) {
	transport, err := newMAXTransport()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout:   defaultHTTPTimeout,
		Transport: transport,
	}

	return maxbot.New(token, maxbot.WithHTTPClient(httpClient))
}

// NewServiceHTTPClient - клиент к API ArcticPass
func NewServiceHTTPClient() (*http.Client, error) {
	transport, err := newMAXTransport()
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp4", addr)
	}
	transport.TLSHandshakeTimeout = 15 * time.Second
	transport.ResponseHeaderTimeout = 20 * time.Second
	return &http.Client{Timeout: 30 * time.Second, Transport: transport}, nil
}

// PrintBot - запрашивает карточку бота. Сбой GetBot не останавливает процесс
func PrintBot(ctx context.Context, api *maxbot.Api) {
	FetchBotInfo(ctx, api)
}

// FetchBotInfo - карточка бота для логов и настройки приложения
func FetchBotInfo(ctx context.Context, api *maxbot.Api) *schemes.BotInfo {
	info, err := api.Bots.GetBot(ctx)
	if err != nil {
		log.Printf("Предупреждение GetBot: %v", err)
		return nil
	}
	fmt.Printf("Get me: %#v\n", info)
	return info
}

// newMAXTransport - HTTP
func newMAXTransport() (*http.Transport, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}

	if !pool.AppendCertsFromPEM(embeddedRussianTrustedCA) {
		return nil, fmt.Errorf("встроенный CA MAX: не удалось прочитать certs/russian_trusted_ca.pem")
	}

	if path := strings.TrimSpace(os.Getenv("MAX_CA_CERT")); path != "" {
		pem, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("MAX_CA_CERT: read file: %w", err)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("MAX_CA_CERT: no certificates found in %s", path)
		}
	}

	insecure := strings.EqualFold(os.Getenv("MAX_TLS_INSECURE"), "1") ||
		strings.EqualFold(os.Getenv("MAX_TLS_INSECURE"), "true")

	if insecure {
		fmt.Println("Предупреждение: MAX_TLS_INSECURE включён - проверка TLS отключена (только для разработки)")
	}

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		TLSHandshakeTimeout: 15 * time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: insecure,
			MinVersion:         tls.VersionTLS12,
		},
	}, nil
}
