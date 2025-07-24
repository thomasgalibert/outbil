package cmd

import (
	"fmt"
	"os"
	"outbil/models"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func generatePDFMaroto(quote *models.Quote, company *models.Company, filename string) error {
	// Configuration du PDF
	cfg := config.NewBuilder().
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	// En-tête avec infos société et logo
	if company != nil {
		// Colonnes pour l'en-tête
		companyCol := col.New(8)
		companyCol.Add(text.New(company.Name, props.Text{
			Size:  16,
			Style: fontstyle.Bold,
		}))
		
		if company.Address != "" {
			companyCol.Add(text.New(company.Address, props.Text{
				Size: 10,
				Top:  8,
			}))
		}
		if company.City != "" {
			companyCol.Add(text.New(fmt.Sprintf("%s %s", company.PostalCode, company.City), props.Text{
				Size: 10,
				Top:  13,
			}))
		}
		if company.Phone != "" {
			companyCol.Add(text.New(fmt.Sprintf("Tél: %s", company.Phone), props.Text{
				Size: 10,
				Top:  18,
			}))
		}
		if company.Email != "" {
			companyCol.Add(text.New(fmt.Sprintf("Email: %s", company.Email), props.Text{
				Size: 10,
				Top:  23,
			}))
		}
		if company.TaxID != "" {
			companyCol.Add(text.New(fmt.Sprintf("N° TVA: %s", company.TaxID), props.Text{
				Size: 10,
				Top:  28,
			}))
		}

		// Colonne pour le logo
		logoCol := col.New(4)
		
		// Vérifier si un logo existe
		logoFiles := []string{"logo.jpg", "logo.jpeg", "logo.png"}
		for _, logoPath := range logoFiles {
			if _, err := os.Stat(logoPath); err == nil {
				img := image.NewFromFile(logoPath, props.Rect{
					Left:   0,
					Top:    0,
					Percent: 80,
					Center: true,
				})
				logoCol.Add(img)
				break
			}
		}

		m.AddRow(35, companyCol, logoCol)
	}

	// Titre DEVIS
	m.AddRow(15,
		col.New(12).Add(
			text.New("DEVIS", props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)

	// Numéro de devis
	yearMonth := quote.Date.Format("2006-01")
	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("%s-%s", yearMonth, quote.QuoteNumber), props.Text{
				Size:   12,
				Family: "Courier",
				Align:  align.Center,
			}),
		),
	)

	// Dates
	m.AddRow(8,
		col.New(6).Add(
			text.New(fmt.Sprintf("Date: %s", quote.Date.Format("02/01/2006")), props.Text{
				Size: 10,
			}),
		),
		col.New(6).Add(
			text.New(fmt.Sprintf("Valable jusqu'au: %s", quote.ValidUntil.Format("02/01/2006")), props.Text{
				Size:  10,
				Align: align.Right,
			}),
		),
	)

	// Espace
	m.AddRow(10)

	// Infos client
	m.AddRow(8,
		col.New(12).Add(
			text.New("CLIENT:", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
			}),
		),
	)

	if quote.Client != nil {
		m.AddRow(5,
			col.New(12).Add(
				text.New(quote.Client.Name, props.Text{
					Size: 10,
				}),
			),
		)

		if quote.Client.Company != "" && quote.Client.Company != quote.Client.Name {
			m.AddRow(5,
				col.New(12).Add(
					text.New(quote.Client.Company, props.Text{
						Size: 10,
					}),
				),
			)
		}

		if quote.Client.Address != "" {
			m.AddRow(5,
				col.New(12).Add(
					text.New(quote.Client.Address, props.Text{
						Size: 10,
					}),
				),
			)
		}

		m.AddRow(5,
			col.New(12).Add(
				text.New(fmt.Sprintf("%s %s", quote.Client.PostalCode, quote.Client.City), props.Text{
					Size: 10,
				}),
			),
		)

		if quote.Client.TaxID != "" {
			m.AddRow(5,
				col.New(12).Add(
					text.New(fmt.Sprintf("N° TVA: %s", quote.Client.TaxID), props.Text{
						Size: 10,
					}),
				),
			)
		}
	}

	// Espace avant tableau
	m.AddRow(10)

	// En-tête du tableau
	headerStyle := &props.Cell{
		BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		BorderType:      border.Full,
		BorderThickness: 0.5,
	}

	m.AddRow(10,
		col.New(5).Add(
			text.New("Description", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Left,
				Left:  2,
			}),
		).WithStyle(headerStyle),
		col.New(1).Add(
			text.New("Qté", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		).WithStyle(headerStyle),
		col.New(2).Add(
			text.New("PU HT", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Right,
			}),
		).WithStyle(headerStyle),
		col.New(1).Add(
			text.New("TVA", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		).WithStyle(headerStyle),
		col.New(3).Add(
			text.New("Total HT", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Right,
			}),
		).WithStyle(headerStyle),
	)

	// Style pour les cellules du tableau
	cellStyle := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.5,
	}

	// Lignes du tableau
	for _, item := range quote.Items {
		// Calculer la hauteur en fonction du texte
		lines := len(item.Description) / 40 // approximation
		if lines < 1 {
			lines = 1
		}
		rowHeight := float64(8 + lines*4)
		
		m.AddRow(rowHeight,
			col.New(5).Add(
				text.New(item.Description, props.Text{
					Size:  9,
					Align: align.Left,
					Left:  2,
					Top:   2,
				}),
			).WithStyle(cellStyle),
			col.New(1).Add(
				text.New(fmt.Sprintf("%.2f", item.Quantity), props.Text{
					Size:   9,
					Align:  align.Center,
					Family: "Courier",
					Top:    2,
				}),
			).WithStyle(cellStyle),
			col.New(2).Add(
				text.New(fmt.Sprintf("%.2f", item.UnitPrice), props.Text{
					Size:   9,
					Align:  align.Right,
					Family: "Courier",
					Top:    2,
				}),
			).WithStyle(cellStyle),
			col.New(1).Add(
				text.New(fmt.Sprintf("%.0f%%", item.TaxRate), props.Text{
					Size:   9,
					Align:  align.Center,
					Family: "Courier",
					Top:    2,
				}),
			).WithStyle(cellStyle),
			col.New(3).Add(
				text.New(fmt.Sprintf("%.2f", item.Amount), props.Text{
					Size:   9,
					Align:  align.Right,
					Family: "Courier",
					Top:    2,
				}),
			).WithStyle(cellStyle),
		)
	}

	// Espace
	m.AddRow(10)

	// Totaux
	subtotal := quote.TotalAmount - quote.TaxAmount + quote.Discount

	// Sous-total
	m.AddRow(6,
		col.New(8),
		col.New(2).Add(
			text.New("Sous-total HT:", props.Text{
				Size:  10,
				Align: align.Right,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%.2f EUR", subtotal), props.Text{
				Size:   10,
				Align:  align.Right,
				Family: "Courier",
			}),
		),
	)

	// Remise
	if quote.Discount > 0 {
		m.AddRow(6,
			col.New(8),
			col.New(2).Add(
				text.New("Remise:", props.Text{
					Size:  10,
					Align: align.Right,
				}),
			),
			col.New(2).Add(
				text.New(fmt.Sprintf("%.2f EUR", quote.Discount), props.Text{
					Size:   10,
					Align:  align.Right,
					Family: "Courier",
				}),
			),
		)
	}

	// TVA
	m.AddRow(6,
		col.New(8),
		col.New(2).Add(
			text.New("TVA:", props.Text{
				Size:  10,
				Align: align.Right,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%.2f EUR", quote.TaxAmount), props.Text{
				Size:   10,
				Align:  align.Right,
				Family: "Courier",
			}),
		),
	)

	// Ligne de séparation
	m.AddRow(2,
		col.New(8),
		col.New(4).Add(
			line.New(props.Line{
				Thickness: 0.5,
			}),
		),
	)

	// Total TTC
	m.AddRow(8,
		col.New(8),
		col.New(2).Add(
			text.New("TOTAL TTC:", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
				Align: align.Right,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%.2f EUR", quote.TotalAmount), props.Text{
				Size:   12,
				Style:  fontstyle.Bold,
				Align:  align.Right,
				Family: "Courier",
			}),
		),
	)

	// Espace
	m.AddRow(15)

	// Notes
	if quote.Notes != "" {
		m.AddRow(6,
			col.New(12).Add(
				text.New("Notes:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
		)
		m.AddRow(0,
			col.New(12).Add(
				text.New(quote.Notes, props.Text{
					Size: 9,
				}),
			),
		)
		m.AddRow(10)
	}

	// Conditions
	if quote.Terms != "" {
		m.AddRow(6,
			col.New(12).Add(
				text.New("Conditions:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
		)
		m.AddRow(0,
			col.New(12).Add(
				text.New(quote.Terms, props.Text{
					Size: 9,
				}),
			),
		)
	}

	// Générer le PDF
	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("erreur lors de la génération du PDF: %w", err)
	}

	// Sauvegarder le PDF temporairement
	tempFile := filename + ".temp"
	err = document.Save(tempFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde du PDF: %w", err)
	}

	// Vérifier si cgv.pdf existe et fusionner si nécessaire
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