package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"outbil/db"
	"outbil/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbSwitchCmd)
	dbCmd.AddCommand(dbCurrentCmd)
	dbCmd.AddCommand(dbDeleteCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Gérer les bases de données",
	Long:  "Gérer plusieurs bases de données (production, demo, test, etc.)",
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lister toutes les bases de données disponibles",
	Run: func(cmd *cobra.Command, args []string) {
		baseDir := filepath.Join(os.Getenv("HOME"), ".outbil")
		
		// Lire le fichier de configuration pour connaître la base active
		currentDB := utils.GetCurrentDatabase()
		
		// Lister tous les fichiers .db
		files, err := filepath.Glob(filepath.Join(baseDir, "*.db"))
		if err != nil {
			utils.Error("Erreur lors de la lecture des bases: %v", err)
			return
		}
		
		if len(files) == 0 {
			utils.Warning("Aucune base de données trouvée")
			utils.Info("Créez-en une avec: outbil db create <nom>")
			return
		}
		
		utils.Info("Bases de données disponibles:")
		for _, file := range files {
			dbName := strings.TrimSuffix(filepath.Base(file), ".db")
			if dbName == currentDB {
				utils.Success("  → %s (active)", dbName)
			} else {
				fmt.Printf("    %s\n", dbName)
			}
			
			// Afficher la taille
			if info, err := os.Stat(file); err == nil {
				size := info.Size()
				if size < 1024 {
					fmt.Printf("      Taille: %d octets\n", size)
				} else if size < 1024*1024 {
					fmt.Printf("      Taille: %.1f Ko\n", float64(size)/1024)
				} else {
					fmt.Printf("      Taille: %.1f Mo\n", float64(size)/(1024*1024))
				}
			}
		}
	},
}

var dbCreateCmd = &cobra.Command{
	Use:   "create [nom]",
	Short: "Créer une nouvelle base de données",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbName := args[0]
		
		// Valider le nom
		if strings.Contains(dbName, "/") || strings.Contains(dbName, "\\") || strings.Contains(dbName, ".") {
			utils.Error("Le nom de base ne doit pas contenir de caractères spéciaux")
			return
		}
		
		baseDir := filepath.Join(os.Getenv("HOME"), ".outbil")
		dbPath := filepath.Join(baseDir, dbName+".db")
		
		// Vérifier si elle existe déjà
		if _, err := os.Stat(dbPath); err == nil {
			utils.Error("La base de données '%s' existe déjà", dbName)
			return
		}
		
		// Créer le répertoire si nécessaire
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			utils.Error("Impossible de créer le répertoire: %v", err)
			return
		}
		
		// Créer et initialiser la nouvelle base
		database, err := db.New(dbPath)
		if err != nil {
			utils.Error("Impossible de créer la base: %v", err)
			return
		}
		defer database.Close()
		
		utils.Success("Base de données '%s' créée avec succès", dbName)
		
		// Proposer de la définir comme base active
		if utils.GetCurrentDatabase() != dbName {
			utils.Info("\nPour utiliser cette base, exécutez:")
			utils.Info("  outbil db switch %s", dbName)
		}
	},
}

var dbSwitchCmd = &cobra.Command{
	Use:   "switch [nom]",
	Short: "Changer de base de données active",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbName := args[0]
		
		baseDir := filepath.Join(os.Getenv("HOME"), ".outbil")
		dbPath := filepath.Join(baseDir, dbName+".db")
		
		// Vérifier que la base existe
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			utils.Error("La base de données '%s' n'existe pas", dbName)
			utils.Info("Bases disponibles:")
			dbListCmd.Run(cmd, []string{})
			return
		}
		
		// Sauvegarder la configuration
		configPath := filepath.Join(baseDir, "config")
		err := os.WriteFile(configPath, []byte(dbName), 0644)
		if err != nil {
			utils.Error("Impossible de sauvegarder la configuration: %v", err)
			return
		}
		
		utils.Success("Base de données active: %s", dbName)
		
		// Vérifier si la base a une entreprise configurée
		database, err := db.New(dbPath)
		if err == nil {
			defer database.Close()
			_, err = database.GetCompany()
			if err != nil {
				utils.Warning("\nCette base n'a pas encore d'entreprise configurée")
				utils.Info("Configurez-la avec: outbil company setup")
			}
		}
	},
}

var dbCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Afficher la base de données active",
	Run: func(cmd *cobra.Command, args []string) {
		currentDB := utils.GetCurrentDatabase()
		utils.Info("Base de données active: %s", currentDB)
		
		// Afficher le chemin complet
		dbPath := utils.GetDatabasePath()
		utils.Info("Chemin: %s", dbPath)
		
		// Afficher des stats
		if info, err := os.Stat(dbPath); err == nil {
			size := info.Size()
			if size < 1024*1024 {
				utils.Info("Taille: %.1f Ko", float64(size)/1024)
			} else {
				utils.Info("Taille: %.1f Mo", float64(size)/(1024*1024))
			}
			utils.Info("Dernière modification: %s", info.ModTime().Format("02/01/2006 15:04"))
		}
	},
}

var dbDeleteCmd = &cobra.Command{
	Use:   "delete [nom]",
	Short: "Supprimer une base de données",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbName := args[0]
		
		// Empêcher la suppression de la base active
		if dbName == utils.GetCurrentDatabase() {
			utils.Error("Impossible de supprimer la base active")
			utils.Info("Changez d'abord de base avec: outbil db switch <autre-base>")
			return
		}
		
		baseDir := filepath.Join(os.Getenv("HOME"), ".outbil")
		dbPath := filepath.Join(baseDir, dbName+".db")
		
		// Vérifier que la base existe
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			utils.Error("La base de données '%s' n'existe pas", dbName)
			return
		}
		
		// Demander confirmation
		utils.Warning("Êtes-vous sûr de vouloir supprimer la base '%s' ?", dbName)
		utils.Warning("Cette action est irréversible!")
		
		var confirm string
		fmt.Print("Tapez le nom de la base pour confirmer: ")
		fmt.Scanln(&confirm)
		
		if confirm != dbName {
			utils.Info("Suppression annulée")
			return
		}
		
		// Supprimer la base
		err := os.Remove(dbPath)
		if err != nil {
			utils.Error("Erreur lors de la suppression: %v", err)
			return
		}
		
		utils.Success("Base de données '%s' supprimée", dbName)
	},
}