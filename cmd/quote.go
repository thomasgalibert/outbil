package cmd

import (
	"fmt"
	"os"
	"outbil/db"
	"outbil/models"
	"outbil/utils"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(quoteCmd)
	quoteCmd.AddCommand(quoteListCmd)
	quoteCmd.AddCommand(quoteCreateCmd)
	quoteCmd.AddCommand(quoteEditCmd)
	quoteCmd.AddCommand(quoteShowCmd)
	quoteCmd.AddCommand(quoteStatusCmd)
	quoteCmd.AddCommand(quoteDeleteCmd)
	quoteCmd.AddCommand(quotePDFCmd)
	quoteCmd.AddCommand(quoteDuplicateCmd)
}

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Gérer les devis",
	Long:  `Commandes pour créer, lister, modifier et exporter des devis`,
}

var quoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lister tous les devis",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New(utils.GetDatabasePath())
		if err != nil {
			utils.Error("Erreur d'ouverture de la base: %v", err)
			return
		}
		defer database.Close()

		quotes, err := database.ListQuotes()
		if err != nil {
			utils.Error("Erreur lors de la récupération des devis: %v", err)
			return
		}

		if len(quotes) == 0 {
			utils.Info("Aucun devis trouvé")
			return
		}

		table := utils.CreateTable()
		table.Header("ID", "Numéro", "Client", "Date", "Montant", "Statut")

		for _, quote := range quotes {
			statusColor := getStatusColor(quote.Status)
			// Afficher avec le préfixe année-mois
			fullNumber := fmt.Sprintf("%s-%s", quote.Date.Format("2006-01"), quote.QuoteNumber)
			table.Append([]string{
				strconv.Itoa(quote.ID),
				fullNumber,
				quote.Client.Name,
				quote.Date.Format("02/01/2006"),
				utils.FormatPrice(quote.TotalAmount, "EUR"),
				statusColor,
			})
		}

		table.Render()
	},
}

var quoteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Créer un nouveau devis",
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
			utils.Error("Aucun client trouvé. Créez d'abord un client avec 'outbil client add'")
			return
		}

		clientNames := make([]string, len(clients))
		for i, client := range clients {
			clientNames[i] = fmt.Sprintf("%s (%s)", client.Name, client.Company)
		}

		prompt := promptui.Select{
			Label: "Sélectionner un client",
			Items: clientNames,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return
		}

		selectedClient := clients[index]

		quote := &models.Quote{
			ClientID: selectedClient.ID,
			Date:     time.Now(),
			Status:   models.StatusDraft,
		}

		quoteNumber, err := database.GetNextQuoteNumber()
		if err != nil {
			utils.Error("Erreur lors de la génération du numéro: %v", err)
			return
		}
		quote.QuoteNumber = quoteNumber

		validityPrompt := promptui.Prompt{
			Label:   "Durée de validité (jours)",
			Default: "30",
		}
		validityDays, _ := validityPrompt.Run()
		days, _ := strconv.Atoi(validityDays)
		quote.ValidUntil = time.Now().AddDate(0, 0, days)

		notesPrompt := promptui.Prompt{
			Label: "Notes (optionnel)",
		}
		quote.Notes, _ = notesPrompt.Run()

		termsPrompt := promptui.Prompt{
			Label:   "Conditions de paiement",
			Default: "Paiement à 30 jours",
		}
		quote.Terms, _ = termsPrompt.Run()

		utils.Info("Ajout des lignes du devis (tapez 'fin' pour terminer)")

		var items []models.QuoteItem
		itemNumber := 1

		for {
			fmt.Printf("\n--- Ligne %d ---\n", itemNumber)

			descPrompt := promptui.Prompt{
				Label: "Description (ou 'fin' pour terminer)",
			}
			description, _ := descPrompt.Run()

			if strings.ToLower(description) == "fin" {
				break
			}

			item := models.QuoteItem{
				Description: description,
			}

			qtyPrompt := promptui.Prompt{
				Label:   "Quantité",
				Default: "1",
			}
			qtyStr, _ := qtyPrompt.Run()
			item.Quantity, _ = utils.ParseFloat(qtyStr)

			pricePrompt := promptui.Prompt{
				Label: "Prix unitaire HT",
			}
			priceStr, _ := pricePrompt.Run()
			item.UnitPrice, _ = utils.ParseFloat(priceStr)

			taxPrompt := promptui.Prompt{
				Label:   "Taux TVA (%)",
				Default: "20",
			}
			taxStr, _ := taxPrompt.Run()
			item.TaxRate, _ = utils.ParseFloat(taxStr)

			item.Amount = item.Quantity * item.UnitPrice
			items = append(items, item)

			utils.Success("Ligne ajoutée: %.2f x %.2f = %.2f EUR HT",
				item.Quantity, item.UnitPrice, item.Amount)

			itemNumber++
		}

		if len(items) == 0 {
			utils.Error("Aucune ligne ajoutée, création annulée")
			return
		}

		quote.Items = items

		var subtotal, totalTax float64
		for _, item := range items {
			subtotal += item.Amount
			totalTax += item.Amount * item.TaxRate / 100
		}

		quote.TaxAmount = totalTax
		quote.TotalAmount = subtotal + totalTax - quote.Discount

		fmt.Printf("\n--- Récapitulatif ---\n")
		fmt.Printf("Sous-total HT: %.2f EUR\n", subtotal)
		fmt.Printf("TVA:           %.2f EUR\n", totalTax)
		fmt.Printf("Total TTC:     %.2f EUR\n", quote.TotalAmount)

		confirm := promptui.Prompt{
			Label:     "Confirmer la création du devis",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.CreateQuote(quote)
			if err != nil {
				utils.Error("Erreur lors de la création: %v", err)
				return
			}
			fullNumber := fmt.Sprintf("%s-%s", quote.Date.Format("2006-01"), quote.QuoteNumber)
			utils.Success("Devis %s créé avec succès (ID: %d)", fullNumber, quote.ID)
		} else {
			utils.Info("Création annulée")
		}
	},
}

var quoteShowCmd = &cobra.Command{
	Use:   "show [ID]",
	Short: "Afficher les détails d'un devis",
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

		quote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Devis non trouvé: %v", err)
			return
		}

		fullNumber := fmt.Sprintf("%s-%s", quote.Date.Format("2006-01"), quote.QuoteNumber)
		fmt.Printf("\n=== DEVIS %s ===\n", fullNumber)
		fmt.Printf("Date: %s\n", quote.Date.Format("02/01/2006"))
		fmt.Printf("Valide jusqu'au: %s\n", quote.ValidUntil.Format("02/01/2006"))
		fmt.Printf("Statut: %s\n", getStatusColor(quote.Status))

		fmt.Printf("\n--- Client ---\n")
		fmt.Printf("%s\n", quote.Client.Name)
		if quote.Client.Company != "" {
			fmt.Printf("%s\n", quote.Client.Company)
		}
		fmt.Printf("%s\n", quote.Client.Address)
		fmt.Printf("%s %s\n", quote.Client.PostalCode, quote.Client.City)
		if quote.Client.TaxID != "" {
			fmt.Printf("N° TVA: %s\n", quote.Client.TaxID)
		}

		fmt.Printf("\n--- Détail ---\n")
		table := utils.CreateTable()
		table.Header("Description", "Qté", "PU HT", "TVA %", "Total HT")

		for _, item := range quote.Items {
			table.Append([]string{
				item.Description,
				fmt.Sprintf("%.2f", item.Quantity),
				fmt.Sprintf("%.2f", item.UnitPrice),
				fmt.Sprintf("%.0f%%", item.TaxRate),
				fmt.Sprintf("%.2f", item.Amount),
			})
		}
		table.Render()

		subtotal := quote.TotalAmount - quote.TaxAmount + quote.Discount
		fmt.Printf("\nSous-total HT: %.2f EUR\n", subtotal)
		if quote.Discount > 0 {
			fmt.Printf("Remise:        %.2f EUR\n", quote.Discount)
		}
		fmt.Printf("TVA:           %.2f EUR\n", quote.TaxAmount)
		fmt.Printf("TOTAL TTC:     %.2f EUR\n", quote.TotalAmount)

		if quote.Notes != "" {
			fmt.Printf("\nNotes: %s\n", quote.Notes)
		}
		if quote.Terms != "" {
			fmt.Printf("Conditions: %s\n", quote.Terms)
		}
	},
}

var quoteEditCmd = &cobra.Command{
	Use:   "edit [ID]",
	Short: "Modifier un devis existant",
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

		quote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Devis non trouvé: %v", err)
			return
		}

		fullNumber := fmt.Sprintf("%s-%s", quote.Date.Format("2006-01"), quote.QuoteNumber)
		utils.Info("Modification du devis %s", fullNumber)

		menuItems := []string{
			"Modifier le client",
			"Modifier la durée de validité",
			"Modifier les notes",
			"Modifier les conditions de paiement",
			"Modifier les lignes du devis",
			"Terminer les modifications",
		}

		for {
			prompt := promptui.Select{
				Label: "Que souhaitez-vous modifier?",
				Items: menuItems,
			}

			index, _, err := prompt.Run()
			if err != nil {
				return
			}

			switch index {
			case 0: // Modifier le client
				clients, err := database.ListClients()
				if err != nil {
					utils.Error("Erreur lors de la récupération des clients: %v", err)
					continue
				}

				clientNames := make([]string, len(clients))
				for i, client := range clients {
					clientNames[i] = fmt.Sprintf("%s (%s)", client.Name, client.Company)
				}

				clientPrompt := promptui.Select{
					Label: "Sélectionner un nouveau client",
					Items: clientNames,
				}

				idx, _, err := clientPrompt.Run()
				if err != nil {
					continue
				}

				quote.ClientID = clients[idx].ID
				quote.Client = &clients[idx]
				utils.Success("Client modifié: %s", clients[idx].Name)

			case 1: // Modifier la durée de validité
				validityPrompt := promptui.Prompt{
					Label:   "Durée de validité (jours)",
					Default: fmt.Sprintf("%d", int(quote.ValidUntil.Sub(quote.Date).Hours()/24)),
				}
				validityDays, _ := validityPrompt.Run()
				days, _ := strconv.Atoi(validityDays)
				quote.ValidUntil = quote.Date.AddDate(0, 0, days)
				utils.Success("Validité modifiée: %s", quote.ValidUntil.Format("02/01/2006"))

			case 2: // Modifier les notes
				notesPrompt := promptui.Prompt{
					Label:   "Notes",
					Default: quote.Notes,
				}
				quote.Notes, _ = notesPrompt.Run()
				utils.Success("Notes modifiées")

			case 3: // Modifier les conditions
				termsPrompt := promptui.Prompt{
					Label:   "Conditions de paiement",
					Default: quote.Terms,
				}
				quote.Terms, _ = termsPrompt.Run()
				utils.Success("Conditions modifiées")

			case 4: // Modifier les lignes
				editLinesMenu := []string{
					"Ajouter une ligne",
					"Modifier une ligne existante",
					"Supprimer une ligne",
					"Retour",
				}

				for {
					linePrompt := promptui.Select{
						Label: "Gestion des lignes",
						Items: editLinesMenu,
					}

					lineIndex, _, err := linePrompt.Run()
					if err != nil || lineIndex == 3 {
						break
					}

					switch lineIndex {
					case 0: // Ajouter une ligne
						fmt.Printf("\n--- Nouvelle ligne ---\n")

						descPrompt := promptui.Prompt{Label: "Description"}
						description, _ := descPrompt.Run()

						item := models.QuoteItem{
							QuoteID:     quote.ID,
							Description: description,
						}

						qtyPrompt := promptui.Prompt{Label: "Quantité", Default: "1"}
						qtyStr, _ := qtyPrompt.Run()
						item.Quantity, _ = utils.ParseFloat(qtyStr)

						pricePrompt := promptui.Prompt{Label: "Prix unitaire HT"}
						priceStr, _ := pricePrompt.Run()
						item.UnitPrice, _ = utils.ParseFloat(priceStr)

						taxPrompt := promptui.Prompt{Label: "Taux TVA (%)", Default: "20"}
						taxStr, _ := taxPrompt.Run()
						item.TaxRate, _ = utils.ParseFloat(taxStr)

						item.Amount = item.Quantity * item.UnitPrice
						quote.Items = append(quote.Items, item)
						utils.Success("Ligne ajoutée")

					case 1: // Modifier une ligne
						if len(quote.Items) == 0 {
							utils.Warning("Aucune ligne à modifier")
							continue
						}

						itemDescs := make([]string, len(quote.Items))
						for i, item := range quote.Items {
							itemDescs[i] = fmt.Sprintf("%s (%.2f x %.2f = %.2f EUR)",
								item.Description, item.Quantity, item.UnitPrice, item.Amount)
						}

						itemPrompt := promptui.Select{
							Label: "Sélectionner la ligne à modifier",
							Items: itemDescs,
						}

						itemIdx, _, err := itemPrompt.Run()
						if err != nil {
							continue
						}

						item := &quote.Items[itemIdx]

						descPrompt := promptui.Prompt{
							Label:   "Description",
							Default: item.Description,
						}
						item.Description, _ = descPrompt.Run()

						qtyPrompt := promptui.Prompt{
							Label:   "Quantité",
							Default: fmt.Sprintf("%.2f", item.Quantity),
						}
						qtyStr, _ := qtyPrompt.Run()
						item.Quantity, _ = utils.ParseFloat(qtyStr)

						pricePrompt := promptui.Prompt{
							Label:   "Prix unitaire HT",
							Default: fmt.Sprintf("%.2f", item.UnitPrice),
						}
						priceStr, _ := pricePrompt.Run()
						item.UnitPrice, _ = utils.ParseFloat(priceStr)

						taxPrompt := promptui.Prompt{
							Label:   "Taux TVA (%)",
							Default: fmt.Sprintf("%.0f", item.TaxRate),
						}
						taxStr, _ := taxPrompt.Run()
						item.TaxRate, _ = utils.ParseFloat(taxStr)

						item.Amount = item.Quantity * item.UnitPrice
						utils.Success("Ligne modifiée")

					case 2: // Supprimer une ligne
						if len(quote.Items) == 0 {
							utils.Warning("Aucune ligne à supprimer")
							continue
						}

						itemDescs := make([]string, len(quote.Items))
						for i, item := range quote.Items {
							itemDescs[i] = fmt.Sprintf("%s (%.2f EUR)", item.Description, item.Amount)
						}

						itemPrompt := promptui.Select{
							Label: "Sélectionner la ligne à supprimer",
							Items: itemDescs,
						}

						itemIdx, _, err := itemPrompt.Run()
						if err != nil {
							continue
						}

						quote.Items = append(quote.Items[:itemIdx], quote.Items[itemIdx+1:]...)
						utils.Success("Ligne supprimée")
					}
				}

			case 5: // Terminer
				// Recalculer les totaux
				var subtotal, totalTax float64
				for _, item := range quote.Items {
					subtotal += item.Amount
					totalTax += item.Amount * item.TaxRate / 100
				}

				quote.TaxAmount = totalTax
				quote.TotalAmount = subtotal + totalTax - quote.Discount

				fmt.Printf("\n--- Récapitulatif des modifications ---\n")
				fmt.Printf("Client: %s\n", quote.Client.Name)
				fmt.Printf("Validité: %s\n", quote.ValidUntil.Format("02/01/2006"))
				fmt.Printf("Nombre de lignes: %d\n", len(quote.Items))
				fmt.Printf("Total TTC: %.2f EUR\n", quote.TotalAmount)

				confirm := promptui.Prompt{
					Label:     "Confirmer les modifications",
					IsConfirm: true,
				}
				result, _ := confirm.Run()

				if result == "y" {
					err = database.UpdateQuote(quote)
					if err != nil {
						utils.Error("Erreur lors de la mise à jour: %v", err)
						return
					}
					utils.Success("Devis %s mis à jour avec succès", fullNumber)
				} else {
					utils.Info("Modifications annulées")
				}
				return
			}
		}
	},
}

var quoteStatusCmd = &cobra.Command{
	Use:   "status [ID]",
	Short: "Modifier le statut d'un devis",
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

		quote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Devis non trouvé: %v", err)
			return
		}

		statuses := []string{
			models.StatusDraft,
			models.StatusSent,
			models.StatusAccepted,
			models.StatusRejected,
			models.StatusExpired,
		}

		prompt := promptui.Select{
			Label: fmt.Sprintf("Statut actuel: %s. Nouveau statut", quote.Status),
			Items: statuses,
		}

		_, status, err := prompt.Run()
		if err != nil {
			return
		}

		err = database.UpdateQuoteStatus(id, status)
		if err != nil {
			utils.Error("Erreur lors de la mise à jour: %v", err)
			return
		}

		utils.Success("Statut mis à jour: %s", status)
	},
}

var quoteDeleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Supprimer un devis",
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

		quote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Devis non trouvé: %v", err)
			return
		}

		fullNumber := fmt.Sprintf("%s-%s", quote.Date.Format("2006-01"), quote.QuoteNumber)
		utils.Warning("Devis à supprimer: %s - %s (%.2f EUR)",
			fullNumber, quote.Client.Name, quote.TotalAmount)

		confirm := promptui.Prompt{
			Label:     "Confirmer la suppression",
			IsConfirm: true,
		}
		result, _ := confirm.Run()

		if result == "y" {
			err = database.DeleteQuote(id)
			if err != nil {
				utils.Error("Erreur lors de la suppression: %v", err)
				return
			}
			utils.Success("Devis supprimé avec succès")
		} else {
			utils.Info("Suppression annulée")
		}
	},
}

var quotePDFCmd = &cobra.Command{
	Use:   "pdf [ID]",
	Short: "Exporter un devis en PDF",
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

		quote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Devis non trouvé: %v", err)
			return
		}

		company, err := database.GetCompany()
		if err != nil {
			utils.Error("Erreur lors de la récupération des infos société: %v", err)
			return
		}

		// Créer le dossier quotes s'il n'existe pas
		err = os.MkdirAll("quotes", 0755)
		if err != nil {
			utils.Error("Erreur lors de la création du dossier quotes: %v", err)
			return
		}

		filename := fmt.Sprintf("quotes/%d_%02d_devis_%s.pdf", quote.CreatedAt.Year(), quote.CreatedAt.Month(), quote.QuoteNumber)
		err = generatePDFMaroto(quote, company, filename)
		if err != nil {
			utils.Error("Erreur lors de la génération du PDF: %v", err)
			return
		}

		utils.Success("PDF généré: %s", filename)
	},
}

var quoteDuplicateCmd = &cobra.Command{
	Use:   "duplicate [ID]",
	Short: "Dupliquer un devis existant",
	Long:  "Créer une copie d'un devis existant avec un nouveau numéro et une nouvelle date",
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

		// Récupérer le devis source pour afficher les infos
		sourceQuote, err := database.GetQuote(id)
		if err != nil {
			utils.Error("Impossible de récupérer le devis: %v", err)
			return
		}

		utils.Info("Duplication du devis %s - %s", sourceQuote.QuoteNumber, sourceQuote.Client.Name)
		
		// Demander confirmation
		confirm := promptui.Prompt{
			Label:     "Voulez-vous dupliquer ce devis",
			IsConfirm: true,
		}
		result, _ := confirm.Run()
		
		if result != "y" {
			utils.Info("Duplication annulée")
			return
		}

		// Dupliquer le devis
		newQuote, err := database.DuplicateQuote(id)
		if err != nil {
			utils.Error("Erreur lors de la duplication: %v", err)
			return
		}

		utils.Success("Devis dupliqué avec succès!")
		utils.Info("Nouveau numéro: %s", newQuote.QuoteNumber)
		utils.Info("Date: %s", newQuote.Date.Format("02/01/2006"))
		utils.Info("Validité: %s", newQuote.ValidUntil.Format("02/01/2006"))
		utils.Info("Statut: %s", getStatusColor(newQuote.Status))
		utils.Info("Montant total: %.2f €", newQuote.TotalAmount)
		
		// Proposer d'éditer le nouveau devis
		confirmEdit := promptui.Prompt{
			Label:     "Voulez-vous modifier le nouveau devis",
			IsConfirm: true,
		}
		resultEdit, _ := confirmEdit.Run()
		
		if resultEdit == "y" {
			utils.Info("Vous pouvez maintenant modifier le devis avec: outbil quote edit %d", newQuote.ID)
		}
	},
}

func getStatusColor(status string) string {
	switch status {
	case models.StatusDraft:
		return "📝 Brouillon"
	case models.StatusSent:
		return "📤 Envoyé"
	case models.StatusAccepted:
		return "✅ Accepté"
	case models.StatusRejected:
		return "❌ Refusé"
	case models.StatusExpired:
		return "⏰ Expiré"
	default:
		return status
	}
}
