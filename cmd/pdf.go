package cmd

import (
	"fmt"
	"io"
	"os"
	"outbil/models"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func generatePDF(quote *models.Quote, company *models.Company, filename string) error {
	// Créer le PDF principal du devis
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Activer le support UTF-8
	pdf.SetFont("Arial", "", 12)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Définir les permissions : lecture et impression seulement
	// Permissions : PermPrint = 4 (impression autorisée)
	pdf.SetProtection(4, "", "")

	pdf.AddPage()

	// Vérifier si un logo existe (jpg, jpeg ou png)
	var logoPath string
	var logoType string
	logoFiles := []struct{path, imgType string}{
		{"logo.jpg", "JPG"},
		{"logo.jpeg", "JPG"},
		{"logo.png", "PNG"},
	}
	
	for _, lf := range logoFiles {
		if _, err := os.Stat(lf.path); err == nil {
			logoPath = lf.path
			logoType = lf.imgType
			break
		}
	}
	
	// Insérer le logo s'il existe (ne pas décaler le reste du contenu)
	if logoPath != "" {
		// Obtenir les infos de l'image pour calculer les proportions
		options := gofpdf.ImageOptions{
			ImageType: logoType,
			ReadDpi:   true,
		}
		info := pdf.RegisterImageOptions(logoPath, options)
		if info != nil {
			// Calculer la taille proportionnelle (largeur max 35mm, plus petit)
			maxWidth := 35.0
			ratio := info.Width() / info.Height()
			width := maxWidth
			height := width / ratio
			
			// Si trop haut, limiter la hauteur
			maxHeight := 25.0
			if height > maxHeight {
				height = maxHeight
				width = height * ratio
			}
			
			// Positionner en haut à droite
			x := 210 - 15 - width // A4 = 210mm, marge 15mm
			y := 15.0
			
			pdf.ImageOptions(logoPath, x, y, width, height, false, options, 0, "")
		}
	}

	// Les infos société commencent normalement à leur position habituelle
	if company != nil {
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(190, 10, tr(company.Name))
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		if company.Address != "" {
			pdf.Cell(190, 5, tr(company.Address))
			pdf.Ln(5)
		}
		if company.City != "" {
			pdf.Cell(190, 5, tr(fmt.Sprintf("%s %s", company.PostalCode, company.City)))
			pdf.Ln(5)
		}
		if company.Phone != "" {
			pdf.Cell(190, 5, tr(fmt.Sprintf("Tél: %s", company.Phone)))
			pdf.Ln(5)
		}
		if company.Email != "" {
			pdf.Cell(190, 5, tr(fmt.Sprintf("Email: %s", company.Email)))
			pdf.Ln(5)
		}
		if company.TaxID != "" {
			pdf.Cell(190, 5, tr(fmt.Sprintf("N° TVA: %s", company.TaxID)))
			pdf.Ln(5)
		}
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, tr("DEVIS"))
	pdf.Ln(10)

	// Afficher le numéro avec préfixe année-mois en police monospace plus petite
	pdf.SetFont("Courier", "", 12)
	yearMonth := quote.Date.Format("2006-01")
	pdf.Cell(190, 5, fmt.Sprintf("%s-%s", yearMonth, quote.QuoteNumber))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 5, tr(fmt.Sprintf("Date: %s", quote.Date.Format("02/01/2006"))))
	pdf.Cell(95, 5, tr(fmt.Sprintf("Valable jusqu'au: %s", quote.ValidUntil.Format("02/01/2006"))))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 5, tr("CLIENT:"))
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, tr(quote.Client.Name))
	pdf.Ln(5)
	if quote.Client.Company != "" && quote.Client.Company != quote.Client.Name {
		pdf.Cell(190, 5, tr(quote.Client.Company))
		pdf.Ln(5)
	}
	if quote.Client.Address != "" {
		pdf.Cell(190, 5, tr(quote.Client.Address))
		pdf.Ln(5)
	}
	pdf.Cell(190, 5, tr(fmt.Sprintf("%s %s", quote.Client.PostalCode, quote.Client.City)))
	pdf.Ln(5)
	if quote.Client.TaxID != "" {
		pdf.Cell(190, 5, tr(fmt.Sprintf("N° TVA: %s", quote.Client.TaxID)))
		pdf.Ln(5)
	}

	pdf.Ln(10)

	// Tableau des articles avec largeurs ajustées
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(80, 8, tr("Description"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 8, tr("Qté"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, tr("PU HT"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(20, 8, tr("TVA"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, tr("Total HT"), "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	for _, item := range quote.Items {
		// Calcul de la hauteur nécessaire pour la description
		lineHt := 5.0
		lines := pdf.SplitLines([]byte(tr(item.Description)), 80)
		cellHeight := float64(len(lines)) * lineHt
		if cellHeight < 8 {
			cellHeight = 8
		}

		// Position actuelle
		x, y := pdf.GetXY()

		// Dessiner d'abord toutes les cellules sans texte pour avoir les bordures alignées
		pdf.CellFormat(80, cellHeight, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, cellHeight, fmt.Sprintf("%.2f", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, cellHeight, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(20, cellHeight, fmt.Sprintf("%.0f%%", item.TaxRate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, cellHeight, fmt.Sprintf("%.2f", item.Amount), "1", 0, "R", false, 0, "")
		
		// Revenir à la position de départ pour écrire la description
		pdf.SetXY(x, y)
		
		// Écrire la description dans la première cellule (sans bordure car déjà dessinée)
		pdf.MultiCell(80, lineHt, tr(item.Description), "", "L", false)
		
		// Se positionner pour la ligne suivante
		pdf.SetXY(x, y+cellHeight)
	}

	pdf.Ln(5)

	subtotal := quote.TotalAmount - quote.TaxAmount + quote.Discount

	// Totaux alignés correctement
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(115, 5, "")
	pdf.Cell(40, 5, tr("Sous-total HT:"))
	pdf.CellFormat(35, 5, fmt.Sprintf("%.2f EUR", subtotal), "", 0, "R", false, 0, "")
	pdf.Ln(5)

	if quote.Discount > 0 {
		pdf.Cell(115, 5, "")
		pdf.Cell(40, 5, tr("Remise:"))
		pdf.CellFormat(35, 5, fmt.Sprintf("%.2f EUR", quote.Discount), "", 0, "R", false, 0, "")
		pdf.Ln(5)
	}

	pdf.Cell(115, 5, "")
	pdf.Cell(40, 5, tr("TVA:"))
	pdf.CellFormat(35, 5, fmt.Sprintf("%.2f EUR", quote.TaxAmount), "", 0, "R", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(115, 7, "")
	pdf.Cell(40, 7, tr("TOTAL TTC:"))
	pdf.CellFormat(35, 7, fmt.Sprintf("%.2f EUR", quote.TotalAmount), "", 0, "R", false, 0, "")
	pdf.Ln(15)

	if quote.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 5, tr("Notes:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(190, 5, tr(quote.Notes), "", "", false)
		pdf.Ln(5)
	}

	if quote.Terms != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 5, tr("Conditions:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(190, 5, tr(quote.Terms), "", "", false)
	}

	// Sauvegarder le PDF temporairement
	tempFile := filename + ".temp"
	err := pdf.OutputFileAndClose(tempFile)
	if err != nil {
		return err
	}

	// Vérifier si @cgv.pdf existe
	cgvPath := "cgv.pdf"
	if _, err := os.Stat(cgvPath); err == nil {
		// Fusionner avec les CGV
		err = mergePDFs(tempFile, cgvPath, filename)
		os.Remove(tempFile) // Nettoyer le fichier temporaire
		if err != nil {
			return fmt.Errorf("erreur lors de la fusion avec les CGV: %w", err)
		}
	} else {
		// Pas de CGV, renommer le fichier temporaire
		err = os.Rename(tempFile, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func mergePDFs(quotePDF, cgvPDF, outputPDF string) error {
	// Créer un dossier temporaire pour l'opération
	tempDir, err := os.MkdirTemp("", "outbil_merge_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Copier les fichiers dans le dossier temporaire
	tempQuote := filepath.Join(tempDir, "quote.pdf")
	tempCGV := filepath.Join(tempDir, "cgv.pdf")

	if err := copyFile(quotePDF, tempQuote); err != nil {
		return err
	}
	if err := copyFile(cgvPDF, tempCGV); err != nil {
		return err
	}

	// Fusionner les PDFs
	inFiles := []string{tempQuote, tempCGV}
	err = api.MergeCreateFile(inFiles, outputPDF, false, nil)
	if err != nil {
		return fmt.Errorf("erreur lors de la fusion PDF: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
