package cmd

import (
	"fmt"
	"outbil/db"
	"outbil/models"
	"outbil/utils"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientAddCmd)
	clientCmd.AddCommand(clientEditCmd)
	clientCmd.AddCommand(clientDeleteCmd)
	clientCmd.AddCommand(clientShowCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Gérer les clients",
	Long:  `Commandes pour créer, lister, modifier et supprimer des clients`,
}

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lister tous les clients",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		clients, err := database.ListClients()
		if err != nil {
			utils.Error("Erreur lors de la récupération des clients: %v", err)
			return
		}

		if len(clients) == 0 {
			utils.Info("Aucun client trouvé")
			return
		}

		table := utils.CreateTable()
		table.Header("ID", "Nom", "Entreprise", "Email", "Téléphone", "Ville")

		for _, client := range clients {
			table.Append([]string{
				strconv.Itoa(client.ID),
				client.Name,
				client.Company,
				client.Email,
				client.Phone,
				client.City,
			})
		}

		table.Render()
	},
}

var clientAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Ajouter un nouveau client",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		client := &models.Client{}

		prompt := promptui.Prompt{
			Label: "Nom du client",
			Validate: func(input string) error {
				if len(input) < 2 {
					return fmt.Errorf("le nom doit contenir au moins 2 caractères")
				}
				return nil
			},
		}
		client.Name, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Entreprise"}
		client.Company, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Email"}
		client.Email, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Téléphone"}
		client.Phone, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Adresse"}
		client.Address, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Ville"}
		client.City, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Code postal"}
		client.PostalCode, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Pays", Default: "France"}
		client.Country, _ = prompt.Run()

		prompt = promptui.Prompt{Label: "Numéro TVA"}
		client.TaxID, _ = prompt.Run()

		confirm := promptui.Prompt{
			Label:     "Confirmer la création du client",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.CreateClient(client)
			if err != nil {
				utils.Error("Erreur lors de la création du client: %v", err)
				return
			}
			utils.Success("Client créé avec succès (ID: %d)", client.ID)
		} else {
			utils.Info("Création annulée")
		}
	},
}

var clientEditCmd = &cobra.Command{
	Use:   "edit [ID]",
	Short: "Modifier un client existant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			utils.Error("ID invalide: %v", err)
			return
		}

		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		client, err := database.GetClient(id)
		if err != nil {
			utils.Error("Client non trouvé: %v", err)
			return
		}

		prompt := promptui.Prompt{
			Label:   "Nom du client",
			Default: client.Name,
		}
		client.Name, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Entreprise",
			Default: client.Company,
		}
		client.Company, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Email",
			Default: client.Email,
		}
		client.Email, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Téléphone",
			Default: client.Phone,
		}
		client.Phone, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Adresse",
			Default: client.Address,
		}
		client.Address, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Ville",
			Default: client.City,
		}
		client.City, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Code postal",
			Default: client.PostalCode,
		}
		client.PostalCode, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Pays",
			Default: client.Country,
		}
		client.Country, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Numéro TVA",
			Default: client.TaxID,
		}
		client.TaxID, _ = prompt.Run()

		confirm := promptui.Prompt{
			Label:     "Confirmer les modifications",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.UpdateClient(client)
			if err != nil {
				utils.Error("Erreur lors de la mise à jour: %v", err)
				return
			}
			utils.Success("Client mis à jour avec succès")
		} else {
			utils.Info("Modifications annulées")
		}
	},
}

var clientDeleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Supprimer un client",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			utils.Error("ID invalide: %v", err)
			return
		}

		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		client, err := database.GetClient(id)
		if err != nil {
			utils.Error("Client non trouvé: %v", err)
			return
		}

		utils.Warning("Client à supprimer: %s (%s)", client.Name, client.Company)
		
		confirm := promptui.Prompt{
			Label:     "Confirmer la suppression",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.DeleteClient(id)
			if err != nil {
				utils.Error("Erreur lors de la suppression: %v", err)
				return
			}
			utils.Success("Client supprimé avec succès")
		} else {
			utils.Info("Suppression annulée")
		}
	},
}

var clientShowCmd = &cobra.Command{
	Use:   "show [ID]",
	Short: "Afficher les détails d'un client",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			utils.Error("ID invalide: %v", err)
			return
		}

		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		client, err := database.GetClient(id)
		if err != nil {
			utils.Error("Client non trouvé: %v", err)
			return
		}

		fmt.Printf("\n--- Détails du client ---\n")
		fmt.Printf("ID:         %d\n", client.ID)
		fmt.Printf("Nom:        %s\n", client.Name)
		fmt.Printf("Entreprise: %s\n", client.Company)
		fmt.Printf("Email:      %s\n", client.Email)
		fmt.Printf("Téléphone:  %s\n", client.Phone)
		fmt.Printf("Adresse:    %s\n", client.Address)
		fmt.Printf("Ville:      %s\n", client.City)
		fmt.Printf("Code postal: %s\n", client.PostalCode)
		fmt.Printf("Pays:       %s\n", client.Country)
		fmt.Printf("N° TVA:     %s\n", client.TaxID)
		fmt.Printf("Créé le:    %s\n", client.CreatedAt.Format("02/01/2006"))
		fmt.Printf("Modifié le: %s\n", client.UpdatedAt.Format("02/01/2006"))
	},
}