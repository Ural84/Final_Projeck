package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateFormat - формат даты YYYYMMDD (20060102)
const DateFormat = "20060102"

// afterNow возвращает true, если первая дата больше второй (без учёта времени)
func afterNow(date, now time.Time) bool {
	// Сравниваем только даты, без времени
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	nowOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return dateOnly.After(nowOnly)
}

// NextDate вычисляет следующую дату для задачи в соответствии с правилом повторения
// now - время, от которого ищется ближайшая дата
// dstart - исходное время в формате 20060102, от которого начинается отсчёт повторений
// repeat - правило повторения
// Возвращает следующую дату в формате 20060102 и ошибку
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Если правило пустое, возвращаем ошибку
	if len(repeat) == 0 {
		return "", errors.New("правило повторения не может быть пустым")
	}

	// Получаем date из time.Parse(DateFormat, dstart)
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата dstart: %w", err)
	}

	repeat = strings.TrimSpace(repeat)

	// Разбиваем repeat на составляющие
	parts := strings.Split(repeat, " ")

	// Обработка правила "y" - ежегодно
	if repeat == "y" {
		originalDay := date.Day()
		originalMonth := date.Month()
		
		for {
			date = date.AddDate(1, 0, 0)
			// Если была 29 февраля в високосном году, а следующий год не високосный,
			// то переносим на 1 марта
			if originalMonth == time.February && originalDay == 29 {
				if !isLeapYear(date.Year()) {
					date = time.Date(date.Year(), time.March, 1, 0, 0, 0, 0, time.UTC)
				}
			}
			if afterNow(date, now) {
				break
			}
		}
		return date.Format(DateFormat), nil
	}

	// Обработка правила "d <число>" - через указанное число дней
	if len(parts) == 2 && parts[0] == "d" {
		// Конвертируем интервал в число
		interval, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("некорректное число дней: %w", err)
		}

		// Максимально допустимое число равно 400
		if interval < 1 || interval > 400 {
			return "", errors.New("число дней должно быть от 1 до 400")
		}

		// Увеличиваем date на указанное количество дней до тех пор, пока дата не станет больше now
		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				break
			}
		}

		return date.Format(DateFormat), nil
	}

	// Обработка правила "w <дни недели>" - дни недели (1=понедельник, 7=воскресенье)
	if len(parts) == 2 && parts[0] == "w" {
		return nextDateWeekly(now, date, parts[1])
	}

	// Обработка правила "m <дни месяца> [месяцы]" - дни месяца
	if len(parts) >= 2 && parts[0] == "m" {
		monthsStr := ""
		if len(parts) > 2 {
			monthsStr = parts[2]
		}
		return nextDateMonthly(now, date, parts[1], monthsStr)
	}

	// Неизвестное правило
	return "", errors.New("неподдерживаемый формат правила повторения")
}

// nextDateWeekly обрабатывает правило "w <дни недели>" - дни недели (1=понедельник, 7=воскресенье)
func nextDateWeekly(now, startDate time.Time, daysStr string) (string, error) {
	// Создаём массив для допустимых дней недели
	var weekday [8]bool // индексы 1-7 используются

	// Парсим дни недели
	dayStrs := strings.Split(daysStr, ",")
	for _, dayStr := range dayStrs {
		dayStr = strings.TrimSpace(dayStr)
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", fmt.Errorf("некорректный день недели: %s", dayStr)
		}
		if day < 1 || day > 7 {
			return "", fmt.Errorf("день недели должен быть от 1 до 7, получено: %d", day)
		}
		weekday[day] = true
	}

	// Начинаем поиск от max(startDate, now)
	searchStart := startDate
	if now.After(startDate) {
		searchStart = now
	}

	// Ищем ближайший день недели из списка
	// Проверяем до 14 дней вперёд (2 недели)
	for i := 0; i < 14; i++ {
		candidate := searchStart.AddDate(0, 0, i)
		dayOfWeek := int(candidate.Weekday())
		// В Go: Sunday=0, Monday=1, ..., Saturday=6
		// Нам нужно: Monday=1, ..., Sunday=7
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}

		// Проверяем, подходит ли день недели
		if weekday[dayOfWeek] && afterNow(candidate, now) {
			return candidate.Format(DateFormat), nil
		}
	}

	return "", errors.New("не удалось найти следующую дату в пределах 2 недель")
}

// nextDateMonthly обрабатывает правило "m <дни месяца> [месяцы]"
func nextDateMonthly(now, startDate time.Time, daysStr, monthsStr string) (string, error) {
	// Создаём массивы для допустимых дней и месяцев
	var dayArray [32]bool // индексы 1-31 используются для обычных дней, -1 и -2 обрабатываются отдельно
	var monthArray [13]bool // индексы 1-12 используются

	// Парсим дни месяца
	dayStrs := strings.Split(daysStr, ",")
	hasNegativeDays := false
	for _, dayStr := range dayStrs {
		dayStr = strings.TrimSpace(dayStr)
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", fmt.Errorf("некорректный день месяца: %s", dayStr)
		}
		if day < -2 || day == 0 || day > 31 {
			return "", fmt.Errorf("день месяца должен быть от -2 до -1 или от 1 до 31, получено: %d", day)
		}
		if day > 0 {
			dayArray[day] = true
		} else {
			hasNegativeDays = true
		}
	}

	// Парсим месяцы (опционально)
	if len(monthsStr) > 0 {
		monthStrs := strings.Split(monthsStr, ",")
		for _, monthStr := range monthStrs {
			monthStr = strings.TrimSpace(monthStr)
			m, err := strconv.Atoi(monthStr)
			if err != nil {
				return "", fmt.Errorf("некорректный месяц: %s", monthStr)
			}
			if m < 1 || m > 12 {
				return "", fmt.Errorf("месяц должен быть от 1 до 12, получено: %d", m)
			}
			monthArray[m] = true
		}
	} else {
		// Если месяцы не указаны, все месяцы допустимы
		for i := 1; i <= 12; i++ {
			monthArray[i] = true
		}
	}

	// Начинаем поиск от max(startDate, now)
	searchStart := startDate
	if now.After(startDate) {
		searchStart = now
	}

	// Ищем ближайшую дату, проверяя день за днём
	// Проверяем до 2 лет вперёд (730 дней)
	for i := 0; i < 730; i++ {
		candidate := searchStart.AddDate(0, 0, i)
		
		// Проверяем месяц
		candidateMonth := int(candidate.Month())
		if !monthArray[candidateMonth] {
			continue
		}

		// Проверяем день
		candidateDay := candidate.Day()
		dayMatches := false
		
		if dayArray[candidateDay] {
			dayMatches = true
		} else if hasNegativeDays {
			// Проверяем -1 (последний день) и -2 (предпоследний день)
			daysInMonth := daysInMonth(candidate.Year(), candidate.Month())
			if candidateDay == daysInMonth {
				// Это последний день месяца, проверяем -1
				for _, dayStr := range dayStrs {
					if dayStr == "-1" {
						dayMatches = true
						break
					}
				}
			} else if candidateDay == daysInMonth-1 {
				// Это предпоследний день месяца, проверяем -2
				for _, dayStr := range dayStrs {
					if dayStr == "-2" {
						dayMatches = true
						break
					}
				}
			}
		}

		// Если день и месяц подходят, и дата больше now и не раньше startDate
		if dayMatches && afterNow(candidate, now) && !candidate.Before(startDate) {
			return candidate.Format(DateFormat), nil
		}
	}

	return "", errors.New("не удалось найти следующую дату в пределах 2 лет")
}

// daysInMonth возвращает количество дней в месяце
func daysInMonth(year int, month time.Month) int {
	// Первый день следующего месяца минус один день
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}

// isLeapYear проверяет, является ли год високосным
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
