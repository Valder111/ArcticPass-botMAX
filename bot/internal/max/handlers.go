package max

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"ArcticPassBot/internal/api"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

var objectClient *api.ObjectClient

// SetObjectClient подключает тестовый клиент карточки здания.
func SetObjectClient(c *api.ObjectClient) {
	objectClient = c
}

const (
	payloadAboutNorilsk = "about_norilsk"
	payloadArcticPass   = "arcticpass"
	payloadMenu         = "menu"
	payloadObjectInfo   = "object_info" // tmp-payload для получения информации о здании по api-серверу

	welcomeText = "👋 Привет! Это чат-бот ArcticPassBot. Здесь ты можешь узнать всё о нашем арктическом городе Норильске. Нажми кнопку ниже: ⬇️"
)

// HandleBotStarted - пользователь только что открыл бота.
func HandleBotStarted(ctx context.Context, api *maxbot.Api, update *schemes.BotStartedUpdate) {
	if err := sendMainMenu(ctx, api, update.GetUserID(), update.GetChatID(), welcomeText); err != nil {
		log.Println("Ошибка отправки:", err)
	}
}

// HandleMessageCreated  обычное сообщение или команда (/start, /menu).
// Обычный текст сразу ищет дом: флаг «ждём улицу» в памяти после деплоя всё равно пропал бы.
func HandleMessageCreated(ctx context.Context, api *maxbot.Api, update *schemes.MessageCreatedUpdate) {
	text := strings.TrimSpace(update.Message.Body.Text)

	var reply *maxbot.Message
	if strings.HasPrefix(text, "/") {
		reply = commandReply(api, text)
	} else if text == "" {
		reply = maxbot.NewMessage().
			SetText("📋 Напиши улицу, например Бегичева, или открой меню командой /menu").
			AddKeyboard(mainMenuKeyboard(api))
	} else {
		reply, _ = objectMessage(ctx, api, text)
		reply.AddKeyboard(mainMenuKeyboard(api))
	}

	setRecipient(reply, &update.Message)
	if err := api.Messages.Send(ctx, reply); err != nil {
		log.Println("Ошибка отправки:", err)
	}
}

func commandReply(api *maxbot.Api, text string) *maxbot.Message {
	args := strings.Fields(text)
	command := args[0]

	switch command {
	case "/start":
		return maxbot.NewMessage().SetText(welcomeText).AddKeyboard(mainMenuKeyboard(api))
	case "/menu":
		return maxbot.NewMessage().SetText("🏠 Главное меню:").AddKeyboard(mainMenuKeyboard(api))
	default:
		return maxbot.NewMessage().SetText("🤔 Неизвестная команда: " + text + " 😅")
	}
}

// HandleMessageCallback - пользователь нажал inline-кнопку под сообщением.
func HandleMessageCallback(ctx context.Context, api *maxbot.Api, update *schemes.MessageCallbackUpdate) {
	payload := update.Callback.Payload

	if payload == payloadObjectInfo {
		msg := maxbot.NewMessage().SetText("🏠 Напиши улицу, например Бегичева. Можно сразу с домом: Бегичева 45")
		if err := replaceCallbackMessage(ctx, api, update, msg); err != nil {
			log.Println("Ошибка отправки:", err)
		}
		return
	}

	if step, ok := historyStepFromPayload(payload); ok {
		msg := maxbot.NewMessage().SetText(historyText(step) + " 📖").AddKeyboard(historyKeyboard(api, step))
		if err := replaceCallbackMessage(ctx, api, update, msg); err != nil {
			log.Println("Ошибка отправки:", err)
		}
		return
	}

	if step, ok := passStepFromPayload(payload); ok {
		msg := maxbot.NewMessage().SetText(passText(step) + " ❄️").AddKeyboard(passKeyboard(api, step))
		if err := replaceCallbackMessage(ctx, api, update, msg); err != nil {
			log.Println("Ошибка отправки:", err)
		}
		return
	}

	text := callbackResponse(payload)
	if text == "" {
		text = "😕 Неизвестная кнопка"
	}

	msg := maxbot.NewMessage().SetText(text)
	if payload == payloadMenu {
		msg.AddKeyboard(mainMenuKeyboard(api))
	}
	if err := replaceCallbackMessage(ctx, api, update, msg); err != nil {
		log.Println("Ошибка отправки:", err)
	}
}

// replaceCallbackMessage сначала присылает следующий шаг и только потом убирает текущий.
// Если отправка не удалась или процесс остановился на деплое, предыдущее сообщение остаётся в чате.
func replaceCallbackMessage(ctx context.Context, api *maxbot.Api, update *schemes.MessageCallbackUpdate, msg *maxbot.Message) error {
	setRecipient(msg, update.Message)
	if err := api.Messages.Send(ctx, msg); err != nil {
		return err
	}
	if update.Message != nil && update.Message.Body.Mid != "" {
		if _, err := api.Messages.DeleteMessage(ctx, update.Message.Body.Mid); err != nil {
			log.Println("Удаление сообщения:", err)
		}
	}
	return nil
}

// historyStepFromPayload читает номер шага из кнопки. Старые кнопки без номера открывают начало истории.
func historyStepFromPayload(payload string) (int, bool) {
	switch payload {
	case payloadAboutNorilsk, payloadHistoryPrev:
		return 0, true
	case payloadHistoryNext:
		if len(norilskHistory) > 1 {
			return 1, true
		}
		return 0, true
	default:
		return parseIndexedStep(payload, "norilsk:", len(norilskHistory))
	}
}

func passStepFromPayload(payload string) (int, bool) {
	switch payload {
	case payloadArcticPass, payloadPassPrev:
		return 0, true
	case payloadPassNext:
		if len(arcticPassStory) > 1 {
			return 1, true
		}
		return 0, true
	default:
		return parseIndexedStep(payload, "arcticpass:", len(arcticPassStory))
	}
}

func parseIndexedStep(payload, prefix string, n int) (int, bool) {
	if n == 0 {
		return 0, false
	}
	rest, ok := strings.CutPrefix(payload, prefix)
	if !ok || rest == "" {
		return 0, false
	}
	step, err := strconv.Atoi(rest)
	if err != nil || step < 0 {
		return 0, false
	}
	if step >= n {
		step = n - 1
	}
	return step, true
}

func parseObjectIDFromQR(raw string) (uint, bool) {
	raw = strings.TrimSpace(raw)
	if rest, ok := strings.CutPrefix(strings.ToLower(raw), "arcticpass:object:"); ok {
		raw = rest
	}
	id, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func addLine(b *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fmt.Fprintf(b, "%s: %s\n", label, value)
}

// mainMenuKeyboard - набор кнопок главного меню (колбэк)
func addOpenApp(kb *maxbot.Keyboard) {
	if miniAppOpenAppReady() {
		kb.AddRow().AddOpenApp("🌐 Открыть приложение", miniAppWebApp, miniAppStartParam, miniAppBotUserID)
	}
}

func mainMenuKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	addOpenApp(kb)
	kb.AddRow().
		AddCallback("🗺 О Норильске", schemes.POSITIVE, payloadAboutNorilsk).
		AddCallback("❄️ Как работает", schemes.DEFAULT, payloadArcticPass)
	kb.AddRow().
		AddCallback("🔍 Найти дом по улице", schemes.DEFAULT, payloadObjectInfo)
	return kb
}

func historyKeyboard(api *maxbot.Api, step int) *maxbot.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	row := kb.AddRow()
	if step > 0 {
		row.AddCallback("⬅️ Назад", schemes.DEFAULT, historyPayload(step-1))
	}
	if step < len(norilskHistory)-1 {
		row.AddCallback("Далее ➡️", schemes.POSITIVE, historyPayload(step+1))
	}
	addOpenApp(kb)
	kb.AddRow().AddCallback("🏠 В меню", schemes.DEFAULT, payloadMenu)
	return kb
}

func passKeyboard(api *maxbot.Api, step int) *maxbot.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	row := kb.AddRow()
	if step > 0 {
		row.AddCallback("⬅️ Назад", schemes.DEFAULT, passPayload(step-1))
	}
	if step < len(arcticPassStory)-1 {
		row.AddCallback("Далее ➡️", schemes.POSITIVE, passPayload(step+1))
	}
	addOpenApp(kb)
	kb.AddRow().AddCallback("🏠 В меню", schemes.DEFAULT, payloadMenu)
	return kb
}

// callbackResponse - по внутреннему id кнопки возвращает текст ответа.
func callbackResponse(payload string) string {
	switch payload {
	case payloadMenu:
		return "👇 Выберите раздел на клавиатуре ниже:"
	default:
		return ""
	}
}

// sendMainMenu - одно сообщение с текстом и клавиатурой в личку или чат.
func sendMainMenu(ctx context.Context, api *maxbot.Api, userID, chatID int64, text string) error {
	msg := maxbot.NewMessage().
		SetText(text).
		AddKeyboard(mainMenuKeyboard(api))
	if chatID != 0 {
		msg.SetChat(chatID)
	} else if userID != 0 {
		msg.SetUser(userID)
	}
	return api.Messages.Send(ctx, msg)
}

// setRecipient - подставляет chat_id или user_id, чтобы ответ ушёл туда же, откуда пришло событие.
func setRecipient(msg *maxbot.Message, origin *schemes.Message) {
	if origin == nil {
		return
	}
	if origin.Recipient.ChatId != 0 {
		msg.SetChat(origin.Recipient.ChatId)
		return
	}
	if origin.Recipient.UserId != 0 {
		msg.SetUser(origin.Recipient.UserId)
		return
	}
	if origin.Sender.UserId != 0 {
		msg.SetUser(origin.Sender.UserId)
	}
}
