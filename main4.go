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

func contains(list []string, item string) bool {
	key := normalize(item)
	for _, it := range list {
		if normalize(it) == key {
			return true
		}
	}
	return false
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

func savePacked(packed []string) error {
	data, err := json.MarshalIndent(packed, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(packedFile, data, 0644)
}

func loadPacked() []string {
	data, err := os.ReadFile(packedFile)
	if err != nil {
		if err := savePacked([]string{}); err != nil {
			fmt.Println("Не удалось создать", packedFile+":", err)
		}
		return []string{}
	}

	var packed []string
	if err := json.Unmarshal(data, &packed); err != nil {
		fmt.Println("Не удалось разобрать", packedFile+", продолжаю с пустым списком.")
		return []string{}
	}
	return packed
}

func loadItems() map[string][]string {
	data, err := os.ReadFile(itemsFile)
	if err != nil {
		fmt.Println("Не удалось прочитать", itemsFile+":", err)
		return nil
	}

	var lists map[string][]string
	if err := json.Unmarshal(data, &lists); err != nil {
		fmt.Println("Не удалось разобрать", itemsFile+":", err)
		return nil
	}
	return lists
}

func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimSpace(line)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("1 — добавить вещь")
		fmt.Println("2 — удалить вещь")
		fmt.Println("3 — показать собранные вещи")
		fmt.Println("4 — проверить день")
		fmt.Println("0 — выход")
		fmt.Print("Выберите действие: ")

		switch readLine(reader) {
		case "1":
			fmt.Println("Введите вещи через запятую:")
			input := readLine(reader)

			packed := loadPacked()
			added := []string{}

			for _, raw := range strings.Split(input, ",") {
				item := normalize(raw)
				if item == "" {
					continue
				}
				if contains(packed, item) {
					fmt.Println("Вещь уже есть в списке:", item)
					continue
				}
				packed = append(packed, item)
				added = append(added, item)
			}

			if len(added) > 0 {
				if err := savePacked(packed); err != nil {
					fmt.Println("Не удалось сохранить список:", err)
				} else {
					fmt.Println("Добавлено:", strings.Join(added, ", "))
				}
			}

		case "2":
			fmt.Println("Какие вещи удалить? Через запятую:")
			input := readLine(reader)

			packed := loadPacked()
			removed := []string{}
			notFound := []string{}

			for _, raw := range strings.Split(input, ",") {
				item := normalize(raw)
				if item == "" {
					continue
				}

				found := false
				filtered := packed[:0]
				for _, it := range packed {
					if normalize(it) == normalize(item) {
						found = true
						removed = append(removed, it)
						continue
					}
					filtered = append(filtered, it)
				}
				packed = filtered

				if !found {
					notFound = append(notFound, item)
				}
			}

			if len(removed) > 0 {
				if err := savePacked(packed); err != nil {
					fmt.Println("Не удалось сохранить список:", err)
				} else {
					fmt.Println("Удалено:", strings.Join(removed, ", "))
				}
			}
			if len(notFound) > 0 {
				fmt.Println("Не найдено в списке:", strings.Join(notFound, ", "))
			}

		case "3":
			packed := loadPacked()
			if len(packed) == 0 {
				fmt.Println("Список собранных вещей пуст.")
			} else {
				fmt.Println("Собранные вещи:")
				for _, item := range packed {
					fmt.Println(" -", item)
				}
			}

		case "4":
			lists := loadItems()
			if lists == nil {
				continue
			}

			fmt.Println("Какой сегодня день?")
			day := normalize(readLine(reader))

			required, exists := lists[day]
			if !exists {
				fmt.Println("Для этого дня список не задан.")
				continue
			}

			packed := loadPacked()
			_, extra, missing := checkItems(required, packed)

			if len(extra) == 0 && len(missing) == 0 {
				fmt.Println("Всё хорошо — вы взяли всё необходимое.")
				continue
			}

			if len(missing) > 0 {
				fmt.Println("Нужно ещё взять:", strings.Join(missing, ", "))
			}
			if len(extra) > 0 {
				fmt.Println("Лишние вещи — можно достать:", strings.Join(extra, ", "))
			}

		case "0":
			return

		default:
			fmt.Println("Неизвестная команда, попробуйте ещё раз.")
		}

		fmt.Println()
	}
}
