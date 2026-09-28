package max

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"ArcticPassBot/internal/api"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func objectMessage(ctx context.Context, bot *maxbot.Api, raw string) (*maxbot.Message, string) {
	if objectClient == nil {
		return maxbot.NewMessage().SetText("⛔️ Клиент API не настроен"), ""
	}
	obj, wait, err := resolveBuilding(ctx, raw)
	if err != nil {
		log.Println("запрос здания:", err)
		return maxbot.NewMessage().SetText("🙈 Не получилось найти дом. Попробуй ещё раз ввести улицу.").SetFormat(schemes.Markdown), "wait"
	}
	if wait != "" {
		return maxbot.NewMessage().SetText(wait).SetFormat(schemes.Markdown), "wait"
	}

	apt, err := objectClient.GetApartment(ctx, obj.ID)
	if err != nil {
		log.Println("жилая карточка:", err)
	}

	msg := maxbot.NewMessage().SetText(formatBuilding(obj, apt)).SetFormat(schemes.Markdown)
	png, err := objectClient.QRImage(ctx, obj.ID)
	if err != nil {
		log.Println("qr:", err)
		return msg, ""
	}
	photo, err := bot.Uploads.UploadPhotoFromReaderWithName(ctx, bytes.NewReader(png), "qr.png")
	if err != nil {
		log.Println("загрузка qr:", err)
		return msg, ""
	}
	msg.AddPhoto(photo)
	return msg, ""
}

func resolveBuilding(ctx context.Context, raw string) (*api.ObjectResponse, string, error) {
	raw = strings.TrimSpace(raw)
	if looksLikeQRCode(raw) {
		obj, err := objectClient.FindByQR(ctx, raw)
		if errors.Is(err, api.ErrNotFound) {
			return nil, "😕 Такой QR не найден. Попробуй написать улицу, например Бегичева 🏠.", nil
		}
		return obj, "", err
	}

	list, total, err := objectClient.Search(ctx, raw)
	if err != nil {
		return nil, "", err
	}
	if total == 0 || len(list) == 0 {
		return nil, "🏚 Дом не найден. Попробуй написать улицу, например Бегичева или Бегичева 45 🌏.", nil
	}
	if total == 1 {
		return &list[0], "", nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "🔎 Нашёл %d домов. Напиши улицу и номер, например Бегичева 45 🏘.\n", total)
	limit := len(list)
	if limit > 6 {
		limit = 6
	}
	for _, obj := range list[:limit] {
		addr := strings.Trim(strings.Join([]string{strings.TrimSpace(obj.Street), strings.TrimSpace(obj.HouseNumber)}, ", "), ", ")
		if addr == "" {
			addr = strings.TrimSpace(obj.Name)
		}
		fmt.Fprintf(&b, "• %s\n", addr)
	}
	return nil, strings.TrimRight(b.String(), "\n"), nil
}

func looksLikeQRCode(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if len(raw) != 36 {
		return false
	}
	for i, ch := range raw {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
				return false
			}
		}
	}
	return true
}

func formatBuilding(obj *api.ObjectResponse, apt *api.ApartmentResponse) string {
	var b strings.Builder
	title := strings.TrimSpace(obj.Name)
	if title == "" {
		title = fmt.Sprintf("🏛 Здание #%d", obj.ID)
	} else {
		title = "🏛 " + title
	}
	fmt.Fprintf(&b, "%s\n", title)

	addr := strings.Trim(strings.Join([]string{strings.TrimSpace(obj.Street), strings.TrimSpace(obj.HouseNumber)}, ", "), ", ")
	addLine(&b, "📍 Адрес", addr)
	addLine(&b, "🌎 Район", obj.District)
	if obj.Levels > 0 {
		fmt.Fprintf(&b, "🏢 Этажей: %d\n", obj.Levels)
	}
	if apt != nil {
		if apt.YearOfEntry > 0 {
			fmt.Fprintf(&b, "📅 Год: %d\n", apt.YearOfEntry)
		}
		if apt.Apartments > 0 {
			fmt.Fprintf(&b, "🚪 Квартир: %d\n", apt.Apartments)
		}
		addLine(&b, "🏗 Стены", apt.TypeOfWalls)
		addLine(&b, "🏢 УК", apt.CompanyName)
	}
	fmt.Fprintf(&b, "\n[🌐 Открыть паспорт](%s)", passportURL(obj.ID))
	return strings.TrimRight(b.String(), "\n")
}

func passportURL(id uint) string {
	return fmt.Sprintf("https://miniapp.xn--80aaxhbzfbdhihi.xn--p1ai/objects/%d", id)
}
