package max

import "fmt"

// Главы истории Норильска для пошагового просмотра в чате.
var norilskHistory = []struct {
	title string
	body  string
}{
	{
		title: "❄️ Наш город",
		body:  "Норильск стоит за Полярным кругом, на Таймыре.\n\n🌙 Зимой здесь долгая ночь.\n☀️ Летом солнце почти не садится.\n🏠 Дома держатся на вечной мерзлоте.",
	},
	{
		title: "🌨️ Как он появился",
		body:  "Раньше по этой тундре кочевали ненцы, долгане и другие народы Севера.\n\nГород вырос в XX веке рядом с рудой и комбинатом. Его строили в очень тяжёлых условиях, и эту историю важно помнить бережно.",
	},
	{
		title: "🏔️ Какой он сейчас",
		body:  "Это дом для людей, которые живут среди ветра, пурги и труб комбината.\n\nРядом плато Путорана, озёра и река Норильская. Город суровый, но свой, родной.",
	},
	{
		title: "🏡 Почему проект именно про него?",
		body:  "Дома здесь живут иначе, чем на большой земле.\n\nArcticPass помогает спокойно узнать свой дом: что это за здание и что о нём уже известно.",
	},
}

const (
	// Старые кнопки без номера шага: после деплоя памяти о шаге уже нет.
	payloadHistoryNext = "norilsk_next"
	payloadHistoryPrev = "norilsk_prev"
)

func historyPayload(step int) string {
	return fmt.Sprintf("norilsk:%d", step)
}

func historyText(step int) string {
	if step < 0 || step >= len(norilskHistory) {
		step = 0
	}
	ch := norilskHistory[step]
	return ch.title + "\n\n" + ch.body + "\n\n" + historyProgress(step)
}

func historyProgress(step int) string {
	return fmt.Sprintf("Шаг %d из %d", step+1, len(norilskHistory))
}
