package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const itemsFile = "items.json"
const packedFile = "packed.json"

func normalize(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// checkItems сравнивает нужные и собранные вещи:
// возвращает совпавшие, лишние и недостающие.
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

// loadItems читает списки нужных вещей по дням из items.json.
func loadItems() (map[string][]string, error) {
	data, err := os.ReadFile(itemsFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать %s: %w", itemsFile, err)
	}
	lists := make(map[string][]string)
	if err := json.Unmarshal(data, &lists); err != nil {
		return nil, fmt.Errorf("не удалось разобрать %s: %w", itemsFile, err)
	}
	return lists, nil
}

// loadPacked читает список собранных вещей из packed.json.
// Если файла нет — создаёт его пустым.
func loadPacked() ([]string, error) {
	data, err := os.ReadFile(packedFile)
	if err != nil {
		if os.IsNotExist(err) {
			packed := []string{}
			if err := savePacked(packed); err != nil {
				return nil, err
			}
			return packed, nil
		}
		return nil, fmt.Errorf("не удалось прочитать %s: %w", packedFile, err)
	}

	var packed []string
	if err := json.Unmarshal(data, &packed); err != nil {
		return nil, fmt.Errorf("не удалось разобрать %s: %w", packedFile, err)
	}
	if packed == nil {
		packed = []string{}
	}
	return packed, nil
}

// savePacked сохраняет список собранных вещей в packed.json.
func savePacked(packed []string) error {
	data, err := json.MarshalIndent(packed, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(packedFile, data, 0644)
}

// splitItems разбирает ввод вида "паспорт, зонт" на отдельные вещи.
func splitItems(input string) []string {
	var result []string
	for _, item := range strings.Split(input, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func contains(packed []string, item string) bool {
	key := normalize(item)
	for _, p := range packed {
		if normalize(p) == key {
			return true
		}
	}
	return false
}

// addItems добавляет новые вещи, пропуская уже собранные.
func addItems(packed []string, items []string) ([]string, int) {
	added := 0
	for _, item := range items {
		if contains(packed, item) {
			fmt.Printf("«%s» уже есть в списке собранных вещей.\n", strings.TrimSpace(item))
			continue
		}
		packed = append(packed, strings.TrimSpace(item))
		added++
	}
	return packed, added
}

// removeItems удаляет перечисленные вещи из списка собранных.
func removeItems(packed []string, items []string) []string {
	for _, item := range items {
		key := normalize(item)
		updated := packed[:0]
		found := false
		for _, p := range packed {
			if normalize(p) == key {
				found = true
				continue
			}
			updated = append(updated, p)
		}
		packed = updated
		if !found {
			fmt.Printf("«%s» нет в списке собранных вещей.\n", strings.TrimSpace(item))
		}
	}
	return packed
}

func printCheckResult(required, packed []string) {
	_, extra, missing := checkItems(required, packed)

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

func main() {
	reader := bufio.NewReader(os.Stdin)

	lists, err := loadItems()
	if err != nil {
		fmt.Println("Ошибка:", err)
		fmt.Println(`Создайте файл items.json, например: {"понедельник": ["паспорт", "телефон", "ключи"]}`)
		return
	}

	packed, err := loadPacked()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	for {
		fmt.Println()
		fmt.Println("Выберите действие:")
		fmt.Println("1 — добавить вещь в список собранных")
		fmt.Println("2 — удалить вещь из списка собранных")
		fmt.Println("3 — показать список собранных вещей")
		fmt.Println("4 — проверить, всё ли собрано на день")
		fmt.Println("0 — выход")
		fmt.Print("> ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("Введите вещи через запятую:")
			input, _ := reader.ReadString('\n')
			items := splitItems(input)
			if len(items) == 0 {
				fmt.Println("Ничего не введено.")
				continue
			}
			packed, added := addItems(packed, items)
			if err := savePacked(packed); err != nil {
				fmt.Println("Не удалось сохранить список:", err)
				continue
			}
			fmt.Printf("Добавлено вещей: %d.\n", added)

		case "2":
			fmt.Println("Введите вещи через запятую:")
			input, _ := reader.ReadString('\n')
			items := splitItems(input)
			if len(items) == 0 {
				fmt.Println("Ничего не введено.")
				continue
			}
			packed = removeItems(packed, items)
			if err := savePacked(packed); err != nil {
				fmt.Println("Не удалось сохранить список:", err)
				continue
			}

		case "3":
			if len(packed) == 0 {
				fmt.Println("Список собранных вещей пуст.")
			} else {
				fmt.Println("Собрано:", strings.Join(packed, ", "))
			}

		case "4":
			fmt.Println("Какой сегодня день?")
			dayInput, _ := reader.ReadString('\n')
			day := normalize(dayInput)

			required, exists := lists[day]
			if !exists {
				fmt.Println("Для этого дня список не задан.")
				continue
			}

			printCheckResult(required, packed)

		case "0":
			return

		default:
			fmt.Println("Неизвестная команда.")
		}
	}
}
