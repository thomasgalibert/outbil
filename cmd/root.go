package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "outbil",
	Short: "OutBil - Gestionnaire de devis professionnel",
	Long: `OutBil est un outil en ligne de commande pour créer et gérer 
des devis professionnels avec une base de données SQLite locale.

Fonctionnalités principales:
  - Gestion des clients
  - Création interactive de devis
  - Export PDF des devis
  - Suivi des statuts des devis`,
}

func Execute() error {
	return rootCmd.Execute()
}