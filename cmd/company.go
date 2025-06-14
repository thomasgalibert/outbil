package cmd

import (
	"fmt"
	"outbil/db"
	"outbil/models"
	"outbil/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(companyCmd)
	companyCmd.AddCommand(companySetupCmd)
	companyCmd.AddCommand(companyShowCmd)
}

var companyCmd = &cobra.Command{
	Use:   "company",
	Short: "Gérer les informations de votre entreprise",
	Long:  `Commandes pour configurer les informations de votre entreprise qui apparaîtront sur les devis`,
}

var companySetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configurer les informations de l'entreprise",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		company, err := database.GetCompany()
		if err != nil {
			utils.Error("Erreur lors de la récupération: %v", err)
			return
		}

		if company == nil {
			company = &models.Company{
				Currency: "EUR",
				TaxRate:  20.0,
			}
		}

		prompt := promptui.Prompt{
			Label:   "Nom de l'entreprise",
			Default: company.Name,
			Validate: func(input string) error {
				if len(input) < 2 {
					return fmt.Errorf("le nom doit contenir au moins 2 caractères")
				}
				return nil
			},
		}
		company.Name, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Email",
			Default: company.Email,
		}
		company.Email, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Téléphone",
			Default: company.Phone,
		}
		company.Phone, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Adresse",
			Default: company.Address,
		}
		company.Address, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Ville",
			Default: company.City,
		}
		company.City, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Code postal",
			Default: company.PostalCode,
		}
		company.PostalCode, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Pays",
			Default: func() string {
				if company.Country != "" {
					return company.Country
				}
				return "France"
			}(),
		}
		company.Country, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Numéro TVA",
			Default: company.TaxID,
		}
		company.TaxID, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Site web",
			Default: company.Website,
		}
		company.Website, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Devise",
			Default: company.Currency,
		}
		company.Currency, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "Taux de TVA par défaut (%)",
			Default: fmt.Sprintf("%.0f", company.TaxRate),
		}
		taxStr, _ := prompt.Run()
		company.TaxRate, _ = utils.ParseFloat(taxStr)

		confirm := promptui.Prompt{
			Label:     "Enregistrer les modifications",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.SaveCompany(company)
			if err != nil {
				utils.Error("Erreur lors de l'enregistrement: %v", err)
				return
			}
			utils.Success("Informations de l'entreprise enregistrées")
		} else {
			utils.Info("Modifications annulées")
		}
	},
}

var companyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Afficher les informations de l'entreprise",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		company, err := database.GetCompany()
		if err != nil {
			utils.Error("Erreur lors de la récupération: %v", err)
			return
		}

		if company == nil {
			utils.Info("Aucune information d'entreprise configurée. Utilisez 'outbil company setup'")
			return
		}

		fmt.Printf("\n--- Informations de l'entreprise ---\n")
		fmt.Printf("Nom:         %s\n", company.Name)
		fmt.Printf("Email:       %s\n", company.Email)
		fmt.Printf("Téléphone:   %s\n", company.Phone)
		fmt.Printf("Adresse:     %s\n", company.Address)
		fmt.Printf("Ville:       %s\n", company.City)
		fmt.Printf("Code postal: %s\n", company.PostalCode)
		fmt.Printf("Pays:        %s\n", company.Country)
		fmt.Printf("N° TVA:      %s\n", company.TaxID)
		fmt.Printf("Site web:    %s\n", company.Website)
		fmt.Printf("Devise:      %s\n", company.Currency)
		fmt.Printf("TVA défaut:  %.0f%%\n", company.TaxRate)
	},
}