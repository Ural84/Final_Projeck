package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Формат даты для работы с задачами
const DateFormat = "20060102"

// Проверяет, что дата больше текущей даты (без учета времени)
func afterNow(date, now time.Time) bool {
	// Берем только дату, без времени
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	nowOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return dateOnly.After(nowOnly)
}

// Вычисляет следующую дату для задачи по правилу повторения
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Проверяем, что правило не пустое
	if len(repeat) == 0 {
		return "", errors.New("правило повторения не может быть пустым")
	}

	// Преобразуем строку даты в формат времени
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата dstart: %w", err)
	}

	// Убираем пробелы в начале и конце
	repeat = strings.TrimSpace(repeat)

	// Разбиваем строку правила на части
	parts := strings.Split(repeat, " ")

	// Правило "y" - каждый год
	if repeat == "y" {
		originalDay := date.Day()
		originalMonth := date.Month()
		
		// Добавляем год, пока дата не станет больше текущей
		for {
			date = date.AddDate(1, 0, 0)
			// Если была 29 февраля, а год не високосный, переносим на 1 марта
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

	// Правило "d <число>" - через указанное число дней
	if len(parts) == 2 && parts[0] == "d" {
		// Преобразуем число дней в целое число
		interval, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("некорректное число дней: %w", err)
		}

		// Проверяем, что число дней от 1 до 400
		if interval < 1 || interval > 400 {
			return "", errors.New("число дней должно быть от 1 до 400")
		}

		// Добавляем дни, пока дата не станет больше текущей
		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				break
			}
		}

		return date.Format(DateFormat), nil
	}

	// Правило "w <дни недели>" - дни недели (1=понедельник, 7=воскресенье)
	if len(parts) == 2 && parts[0] == "w" {
		return nextDateWeekly(now, date, parts[1])
	}

	// Правило "m <дни месяца> [месяцы]" - дни месяца
	if len(parts) >= 2 && parts[0] == "m" {
		monthsStr := ""
		if len(parts) > 2 {
			monthsStr = parts[2]
		}
		return nextDateMonthly(now, date, parts[1], monthsStr)
	}

	// Если правило не распознано, возвращаем ошибку
	return "", errors.New("неподдерживаемый формат правила повторения")
}

// Обрабатывает правило "w <дни недели>" - дни недели (1=понедельник, 7=воскресенье)
func nextDateWeekly(now, startDate time.Time, daysStr string) (string, error) {
	// Массив для хранения допустимых дней недели
	var weekday [8]bool

	// Разбиваем строку с днями недели
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

	// Начинаем поиск с начальной даты или текущей даты (какая больше)
	searchStart := startDate
	if now.After(startDate) {
		searchStart = now
	}

	// Ищем ближайший подходящий день недели
	// Проверяем 14 дней вперед
	for i := 0; i < 14; i++ {
		candidate := searchStart.AddDate(0, 0, i)
		dayOfWeek := int(candidate.Weekday())
		// В Go воскресенье = 0, понедельник = 1, и т.д.
		// Нам нужно: понедельник = 1, воскресенье = 7
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}

		// Если день недели подходит и дата больше текущей
		if weekday[dayOfWeek] && afterNow(candidate, now) {
			return candidate.Format(DateFormat), nil
		}
	}

	return "", errors.New("не удалось найти следующую дату в пределах 2 недель")
}

// Обрабатывает правило "m <дни месяца> [месяцы]"
func nextDateMonthly(now, startDate time.Time, daysStr, monthsStr string) (string, error) {
	// Массивы для хранения допустимых дней и месяцев
	var dayArray [32]bool
	var monthArray [13]bool

	// Разбиваем строку с днями месяца
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

	// Разбиваем строку с месяцами (если указана)
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
		// Если месяцы не указаны, разрешаем все месяцы
		for i := 1; i <= 12; i++ {
			monthArray[i] = true
		}
	}

	// Начинаем поиск с начальной даты или текущей даты (какая больше)
	searchStart := startDate
	if now.After(startDate) {
		searchStart = now
	}

	// Ищем ближайшую подходящую дату
	// Проверяем 730 дней вперед (примерно 2 года)
	for i := 0; i < 730; i++ {
		candidate := searchStart.AddDate(0, 0, i)
		
		// Проверяем, подходит ли месяц
		candidateMonth := int(candidate.Month())
		if !monthArray[candidateMonth] {
			continue
		}

		// Проверяем, подходит ли день
		candidateDay := candidate.Day()
		dayMatches := false
		
		if dayArray[candidateDay] {
			dayMatches = true
		} else if hasNegativeDays {
			// Проверяем -1 (последний день месяца) и -2 (предпоследний день)
			daysInMonth := daysInMonth(candidate.Year(), candidate.Month())
			if candidateDay == daysInMonth {
				// Это последний день месяца
				for _, dayStr := range dayStrs {
					if dayStr == "-1" {
						dayMatches = true
						break
					}
				}
			} else if candidateDay == daysInMonth-1 {
				// Это предпоследний день месяца
				for _, dayStr := range dayStrs {
					if dayStr == "-2" {
						dayMatches = true
						break
					}
				}
			}
		}

		// Если день и месяц подходят, и дата больше текущей
		if dayMatches && afterNow(candidate, now) && !candidate.Before(startDate) {
			return candidate.Format(DateFormat), nil
		}
	}

	return "", errors.New("не удалось найти следующую дату в пределах 2 лет")
}

// Возвращает количество дней в месяце
func daysInMonth(year int, month time.Month) int {
	// Берем первый день следующего месяца и вычитаем один день
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}

// Проверяет, является ли год високосным
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
