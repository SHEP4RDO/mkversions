package mkversions

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// AppMetadata содержит метаданные для обновления
type AppMetadata struct {
	ProductVersion string
	ProgramName    string
	Description    string
	Legal          string
	CompanyName    string
	InternalName   string
}

const (
	rceditURL64 = "https://github.com/electron/rcedit/releases/download/v1.1.1/rcedit-x64.exe"
	rceditURL86 = "https://github.com/electron/rcedit/releases/download/v1.1.1/rcedit-x86.exe"
)

func downloadRcedit() (string, error, bool) {
	// Определяем URL в зависимости от архитектуры системы
	url := rceditURL64
	if runtime.GOARCH == "386" {
		url = rceditURL86
	}

	// Определяем путь для сохранения rcedit
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err), false
	}
	rceditPath := filepath.Join(homeDir, "rcedit.exe")

	// Загружаем файл
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download rcedit: %v", err), false
	}
	defer resp.Body.Close()

	// Создаем файл для сохранения
	out, err := os.Create(rceditPath)
	if err != nil {
		return "", fmt.Errorf("failed to create rcedit file: %v", err), false
	}
	defer out.Close()

	// Копируем содержимое ответа в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save rcedit file: %v", err), false
	}

	return rceditPath, nil, true
}

// copyFile копирует файл из source в destination
func copyFile(source, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// UpdateExeMetadata обновляет метаданные в указанном .exe файле
func (i *Info) UpdateExeMetadata(rcedit, exePath string) error {
	cmd := exec.Command(rcedit, exePath,
		"--set-file-version", i.Version,
		"--set-product-version", i.ProductVersion,
		"--set-version-string", "ProductName", i.ProgramName,
		"--set-version-string", "FileDescription", i.Description,
		"--set-version-string", "LegalCopyright", i.Legal,
		"--set-version-string", "CompanyName", i.CompanyName,
		"--set-version-string", "InternalName", i.ProgramName,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update exe metadata: %v, output: %s", err, string(output))
	}
	return nil
}

func (i *Info) performUpdate(rcedit string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	tempPath := exePath + ".tmp"
	backupPath := exePath + ".bak"
	removeSignalPath := exePath + ".remove"

	// Удаляем старый сигнал удаления, если он существует
	os.Remove(removeSignalPath)

	if err := copyFile(exePath, tempPath); err != nil {
		return fmt.Errorf("failed to copy file to %s: %v", tempPath, err)
	}

	if err := i.UpdateExeMetadata(rcedit, tempPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to update metadata: %v", err)
	}

	if err := os.Rename(exePath, backupPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to backup original file: %v", err)
	}

	if err := os.Rename(tempPath, exePath); err != nil {
		os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to replace original file: %v", err)
	}

	if file, err := os.Create(removeSignalPath); err == nil {
		file.Close()
	} else {
		fmt.Printf("Warning: Failed to create remove signal file: %v\n", err)
	}

	fmt.Println("Update successful! Restarting...")
	time.Sleep(1 * time.Second)

	// Перезапускаем программу
	cmd := exec.Command(exePath, "--tmp-clear")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart program: %v", err)
	}

	os.Exit(0)
	return nil
}

func handleRemoveSignal() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return
	}

	removeSignalPath := exePath + ".remove"
	backupPath := exePath + ".bak"

	if _, err := os.Stat(removeSignalPath); err == nil {
		fmt.Println("Remove signal detected")

		// Попытка удалить файл .bak
		err := removeFileWithRetry(backupPath)
		if err != nil {
			fmt.Printf("Failed to remove backup file: %v\n", err)
			// Если не удалось удалить, попробуем переименовать
			newBackupPath := backupPath + ".old"
			if renameErr := os.Rename(backupPath, newBackupPath); renameErr != nil {
				fmt.Printf("Failed to rename backup file: %v\n", renameErr)
			} else {
				fmt.Printf("Backup file renamed to %s\n", newBackupPath)
			}
		} else {
			fmt.Println("Backup file removed successfully")
		}

		// Удаляем файл сигнала
		if err := os.Remove(removeSignalPath); err != nil {
			fmt.Printf("Failed to remove signal file: %v\n", err)
		} else {
			fmt.Println("Signal file removed successfully")
		}
	}
}

func removeFileWithRetry(filePath string) error {
	maxAttempts := 5
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = os.Remove(filePath)
		if err == nil {
			return nil
		}

		if os.IsPermission(err) {
			os.Chmod(filePath, 0666)
		}

		time.Sleep(time.Duration(attempt*100) * time.Millisecond)
	}

	return err
}
func (i *Info) RunUpdate(rceditPath ...string) {
	handleRemoveSignal()

	var path string
	if len(rceditPath) > 0 {
		path = rceditPath[0]
	}

	localPaths := []string{
		"./rcedit.exe",
		"./rcedit-x64.exe",
		"./rcedit-x86.exe",
		"/usr/local/bin/rcedit-x64",
		"/usr/local/bin/rcedit-x86",
		"C:\\Program Files\\rcedit.exe",
		"C:\\Program Files\\rcedit-x64.exe",
		"C:\\Program Files (x86)\\rcedit-x86.exe",
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--tmp-clear":
			handleRemoveSignal()
			fmt.Println("Temporary files cleared. Exiting...")
			os.Exit(0)

		case "--update":
			var err error
			downloadedRcedit := false

			if path == "" {
				fmt.Println("No rcedit path provided. Searching in local directories and standard paths...")
				for _, p := range localPaths {
					if fileExists(p) {
						path = p
						break
					}
				}

				if path == "" {
					fmt.Println("rcedit not found. Attempting to download...")
					path, err, downloadedRcedit = downloadRcedit()
					if err != nil {
						fmt.Printf("Failed to download rcedit: %v\n", err)
						os.Exit(1)
					}
				}
			} else if !fileExists(path) {
				fmt.Printf("rcedit not found at %s.\n", path)
				os.Exit(1)
			}

			if err = i.performUpdate(path); err != nil {
				fmt.Printf("Update failed: %v\n", err)
				os.Exit(1)
			}

			if downloadedRcedit {
				if err := os.Remove(path); err != nil {
					fmt.Printf("Warning: Failed to remove downloaded rcedit: %v\n", err)
				}
			}
		default:
			fmt.Println("Usage: --update or --tmp-clear")
			os.Exit(1)
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
