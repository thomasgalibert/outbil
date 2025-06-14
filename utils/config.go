package utils

import (
	"os"
	"path/filepath"
)

func GetDatabasePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "outbil.db"
	}

	dataDir := filepath.Join(homeDir, ".outbil")
	os.MkdirAll(dataDir, 0755)

	// Lire la base active depuis le fichier de config
	dbName := GetCurrentDatabase()
	return filepath.Join(dataDir, dbName+".db")
}

func GetCurrentDatabase() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "outbil"
	}

	configPath := filepath.Join(homeDir, ".outbil", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Base par défaut si pas de config
		return "outbil"
	}

	dbName := string(data)
	if dbName == "" {
		return "outbil"
	}
	return dbName
}