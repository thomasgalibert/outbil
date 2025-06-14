package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func Success(format string, args ...interface{}) {
	color.Green("✓ " + fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	color.Red("✗ " + fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	color.Blue("ℹ " + fmt.Sprintf(format, args...))
}

func Warning(format string, args ...interface{}) {
	color.Yellow("⚠ " + fmt.Sprintf(format, args...))
}

func CreateTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	return table
}

func FormatPrice(amount float64, currency string) string {
	if currency == "" {
		currency = "EUR"
	}
	return fmt.Sprintf("%.2f %s", amount, currency)
}

func ParseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", -1)
	return strconv.ParseFloat(s, 64)
}