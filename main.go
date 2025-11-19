package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Media struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	Year    string `json:"year"`
	Comment string `json:"comment,omitempty"`
}
type MediaLibrary struct {
	MediaList []Media `json:"media_list"`
}

func (ml *MediaLibrary) AddMedia(media Media) {
	ml.MediaList = append(ml.MediaList, media)
}
func (ml *MediaLibrary) ShowAll() {
	if len(ml.MediaList) == 0 {
		fmt.Println("Список просмотренных фильмов и сериалов пуст.")
		return
	}
	for i, media := range ml.MediaList {
		fmt.Printf("%d. %s (%s) [%s]\n", i+1, media.Title, media.Year, media.Type)
		if media.Comment != "" {
			fmt.Printf("   Комментарий: %s\n", media.Comment)
		}
	}
}
func (ml *MediaLibrary) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(ml)
}
func (ml *MediaLibrary) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ml)
}
func readInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func main() {
	const filename = "media_library.json"
	var library MediaLibrary
	if err := library.Load(filename); err != nil {
		fmt.Println("Ошибка при загрузке данных:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n--- Меню ---")
		fmt.Println("1. Добавить просмотренный фильм или сериал")
		fmt.Println("2. Посмотреть список просмотренных")
		fmt.Println("3. Выйти и сохранить")
		fmt.Print("Выберите действие: ")

		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			title := readInput("Название: ")
			year := readInput("Год выпуска: ")
			typ := readInput("Тип (фильм/сериал): ")
			comment := readInput("Комментарий (опционально): ")

			if typ != "фильм" && typ != "сериал" {
				fmt.Println("Некорректный тип. Попробуйте снова.")
				continue
			}

			newMedia := Media{
				Title:   title,
				Year:    year,
				Type:    typ,
				Comment: comment,
			}
			library.AddMedia(newMedia)
			fmt.Println("Запись добавлена!")

		case "2":
			fmt.Println("\nСписок просмотренных:")
			library.ShowAll()

		case "3":
			if err := library.Save(filename); err != nil {
				fmt.Println("Ошибка при сохранении данных:", err)
			} else {
				fmt.Println("Данные успешно сохранены. Выход.")
			}
			return

		default:
			fmt.Println("Некорректный выбор. Попробуйте снова.")
		}
	}
}
