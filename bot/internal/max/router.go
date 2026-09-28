// package max - router - разбор входящих апдейтов MAX по типу события.
package max

import (
	"context"
	"log"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

// Dispatch - отдаёт апдейт обработчику: старт бота, сообщение или нажатие кнопки.
// Паника в одном апдейте не останавливает процесс: после деплоя очередь может принести старые события.
func Dispatch(ctx context.Context, api *maxbot.Api, upd schemes.UpdateInterface) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("апдейт не обработан: %v", rec)
		}
	}()

	switch updatePayload := upd.(type) {
	case *schemes.BotStartedUpdate:
		HandleBotStarted(ctx, api, updatePayload)
	case *schemes.MessageCreatedUpdate:
		HandleMessageCreated(ctx, api, updatePayload)
	case *schemes.MessageCallbackUpdate:
		HandleMessageCallback(ctx, api, updatePayload)
	}
}
