package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"

	"github.com/common-nighthawk/go-figure"
	"github.com/schollz/progressbar/v3"
)

func main() {
	myFigure := figure.NewFigure("backuper", "standard", true)
	myFigure.Print()
	// Объявляем флаги.
	scanDir := flag.String("d", "", "Директория для сканирования")
	outputDir := flag.String("o", "", "Директория для копирования найденных файлов (необязательно)")
	flag.Parse()

	// Проверяем, что указана директория для сканирования.
	if *scanDir == "" {
		fmt.Println("Использование: -d <директория для сканирования> [-o <директория для копирования>]")
		return
	}

	var totalSize int64
	var fileStats = make(map[string]struct {
		count int
		size  int64
	})
	var fileList []string

	// Определяем расширения, которые нас интересуют.
	extensions := map[string]bool{
		".doc":  true,
		".docx": true,
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".mkv":  true,
	}

	// Сканирование директории.
	s := spinner.New(spinner.CharSets[21], 100*time.Millisecond) // Build our new spinner
	s.Suffix = " Cканирование файлов..."
	s.Start()
	err := filepath.Walk(*scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Ошибка при доступе к %s: %v\n", path, err)
			return nil
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if extensions[ext] {
				// Обновляем статистику.
				totalSize += info.Size()
				stat := fileStats[ext]
				stat.count++
				stat.size += info.Size()
				fileStats[ext] = stat

				// Сохраняем файл в список для возможного копирования.
				fileList = append(fileList, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Ошибка при сканировании: %v\n", err)
		return
	}
	s.Stop()

	// Выводим общий размер только для файлов с интересующими расширениями.
	fmt.Println("\nРезультаты сканирования:")
	if totalSize/int64(math.Pow(10, 9)) < 1 { // 1 ГБ = 2^30 байт
		fmt.Printf("Общий размер файлов с выбранными расширениями: %.2f MB\n", float64(totalSize)/(1<<20))
	} else {
		fmt.Printf("Общий размер файлов с выбранными расширениями: %.2f GB\n", float64(totalSize)/(1<<30))
	}

	// Выводим статистику по расширениям.
	fmt.Println("\nСтатистика по расширениям:")
	for ext, stat := range fileStats {
		sizeMB := float64(stat.size) / (1024 * 1024)
		fmt.Printf("%s: %d файлов, %.2f MB\n", ext, stat.count, sizeMB)
	}

	// Если директория для копирования указана, запускаем процесс копирования.
	if *outputDir != "" {
		fmt.Println("\nКопирование файлов...")
		copyFiles(fileList, *scanDir, *outputDir)
	}

	fmt.Scan("Нажмите любую кнопку для выхода")
}

// Функция для копирования файла с сохранением структуры директорий.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Создаем директорию, если ее нет.
	err = os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	if err != nil {
		return err
	}

	// Копируем содержимое файла.
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Функция для копирования файлов с прогресс-баром.
func copyFiles(fileList []string, baseDir, outputDir string) {
	// Инициализируем прогресс-бар.
	bar := progressbar.Default(int64(len(fileList)), "Копирование файлов")

	for _, src := range fileList {
		// Определяем относительный путь для сохранения структуры директорий.
		relativePath, _ := filepath.Rel(baseDir, src)
		targetPath := filepath.Join(outputDir, relativePath)

		// Копируем файл.
		if err := copyFile(src, targetPath); err != nil {
			fmt.Printf("Ошибка копирования %s: %v\n", src, err)
		}

		// Обновляем прогресс-бар.
		bar.Add(1)
	}
}
