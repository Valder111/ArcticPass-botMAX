package max

import (
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

var (
	miniAppWebApp     string
	miniAppStartParam string
	miniAppBotUserID  int64
)

func ConfigureMiniApp(webAppFromEnv, startParam string, bot *schemes.BotInfo) {
	miniAppStartParam = strings.TrimSpace(startParam)
	username := strings.TrimPrefix(strings.TrimSpace(webAppFromEnv), "@")
	if bot != nil {
		if username == "" {
			username = strings.TrimSpace(bot.Username)
		}
		miniAppBotUserID = bot.UserId
	}
	miniAppWebApp = username
}

func miniAppOpenAppReady() bool {
	return miniAppWebApp != "" && miniAppBotUserID != 0
}
