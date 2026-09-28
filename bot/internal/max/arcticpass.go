package max

import "fmt"

var arcticPassStory = []struct {
	title string
	body  string
}{
	{
		title: "🤔 Зачем этот бот нужен?",
		body:  "ArcticPass - цифровой паспорт зданий города Норильска. Чтобы житель мог спокойно открыть карточку своего здания и понять, что о нём известно.",
	},
	{
		title: "👥 Для кого предназначен?",
		body:  "Для жителей Норильска и соседних районов. Также, для сотрудников управляющих компаний, которым удобнее оставить запись осмотра прямо у дома.",
	},
	{
		title: "⚙️ Как это работает?",
		body:  "На доме есть QR. Его сканируют в боте или в мини-приложении MAX и видят карточку здания. Подробная работа остаётся на сайте, а здесь - быстрые повседневные действия.",
	},
	{
		title: "🔮 Что дальше?",
		body:  "Сейчас это реестр строений и их карточки. Когда появится управляющая компания-партнёр, к паспорту добавится информация об инспекциях и осмотрах объектов.",
	},
}

const (
	// Старые кнопки без номера шага: после деплоя памяти о шаге уже нет.
	payloadPassNext = "arcticpass_next"
	payloadPassPrev = "arcticpass_prev"
)

func passPayload(step int) string {
	return fmt.Sprintf("arcticpass:%d", step)
}

func passText(step int) string {
	if step < 0 || step >= len(arcticPassStory) {
		step = 0
	}
	ch := arcticPassStory[step]
	return ch.title + "\n\n" + ch.body + "\n\n" + fmt.Sprintf("Шаг %d из %d", step+1, len(arcticPassStory))
}
