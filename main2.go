package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type CheckRequest struct {
	Day   string   `json:"day"`
	Taken []string `json:"taken"`
}

type CheckResponse struct {
	Matched []string `json:"matched"`
	Extra   []string `json:"extra"`
	Missing []string `json:"missing"`
	Message string   `json:"message,omitempty"`
}

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
	lists := map[string][]string{
		"понедельник": {"паспорт", "телефон", "ключи"},
		"среда":       {"телефон", "ключи", "вода"},
		"пятница":     {"паспорт", "телефон", "зонт"},
	}

	// Открываем локальный порт и запускаем сервер.
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Не удалось запустить сервер:", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Используйте POST-запрос", http.StatusMethodNotAllowed)
			return
		}

		var request CheckRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Не удалось прочитать данные запроса", http.StatusBadRequest)
			return
		}

		required, exists := lists[normalize(request.Day)]
		if !exists {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(CheckResponse{
				Message: "Для этого дня список не задан",
			})
			return
		}

		matched, extra, missing := checkItems(required, request.Taken)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(CheckResponse{
			Matched: matched,
			Extra:   extra,
			Missing: missing,
		})
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Println("Ошибка сервера:", err)
		}
	}()
	defer server.Shutdown(context.Background())

	fmt.Println("Какой сегодня день?")
	var day string
	if _, err := fmt.Scanln(&day); err != nil {
		fmt.Println("Не удалось прочитать день.")
		return
	}

	fmt.Println("Введите взятые вещи через запятую:")
	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		fmt.Println("Не удалось прочитать вещи.")
		return
	}

	requestBody, err := json.Marshal(CheckRequest{
		Day:   day,
		Taken: strings.Split(input, ","),
	})
	if err != nil {
		fmt.Println("Не удалось подготовить запрос:", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Post(
		"http://127.0.0.1:8080/check",
		"application/json",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		fmt.Println("Не удалось отправить запрос:", err)
		return
	}
	defer response.Body.Close()

	var result CheckResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		fmt.Println("Не удалось прочитать ответ сервера:", err)
		return
	}

	if response.StatusCode == http.StatusNotFound {
		fmt.Println(result.Message)
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Сервер вернул ошибку:", response.Status)
		return
	}

	if len(result.Extra) == 0 && len(result.Missing) == 0 {
		fmt.Println("Всё хорошо — вы взяли всё необходимое.")
		return
	}
	if len(result.Missing) > 0 {
		fmt.Println("Нужно ещё взять:", strings.Join(result.Missing, ", "))
	}
	if len(result.Extra) > 0 {
		fmt.Println("Лишние вещи — можно достать:", strings.Join(result.Extra, ", "))
	}
}
