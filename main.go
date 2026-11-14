package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func normalize(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

func checkItems(required, taken []string) (matched, extra, missing []string) {
	requiredSet := make(map[string]string)
	takenSet := make(map[string]bool)

	for _, item := range required {
		requiredSet[normalize(item)] = item
	}

	for _, item := range taken {
		key := normalize(item)
		if key == "" || takenSet[key] {
			continue
		}
		takenSet[key] = true

		if requiredItem, exists := requiredSet[key]; exists {
			matched = append(matched, requiredItem)
		} else {
			extra = append(extra, strings.TrimSpace(item))
		}
	}

	for _, item := range required {
		if !takenSet[normalize(item)] {
			missing = append(missing, item)
		}
	}

	return
}

func main() {
	// Для каждого дня задан свой список необходимых вещей.
	lists := map[string][]string{
		"понедельник": {"паспорт", "телефон", "ключи"},
		"среда":       {"телефон", "ключи", "вода"},
		"пятница":     {"паспорт", "телефон", "зонт"},
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Какой сегодня день?")
	dayInput, _ := reader.ReadString('\n')
	day := normalize(dayInput)

	required, exists := lists[day]
	if !exists {
		fmt.Println("Для этого дня список не задан.")
		return
	}

	fmt.Println("Введите взятые вещи через запятую:")
	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		fmt.Println("Не удалось прочитать ввод.")
		return
	}

	taken := strings.Split(input, ",")
	_, extra, missing := checkItems(required, taken)

	if len(extra) == 0 && len(missing) == 0 {
		fmt.Println("Всё хорошо — вы взяли всё необходимое.")
		return
	}

	if len(missing) > 0 {
		fmt.Println("Нужно ещё взять:", strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		fmt.Println("Лишние вещи — можно достать:", strings.Join(extra, ", "))
	}
}
