package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var excelCmd = &cobra.Command{Use: "excel", Short: "XLSX CRUD operations"}

var excelCreateCmd = &cobra.Command{
	Use: "create <file> [sheet]", Short: "Create workbook",
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		book := excelize.NewFile()
		defer book.Close()
		if len(args) == 2 && args[1] != "Sheet1" {
			if err := book.SetSheetName("Sheet1", args[1]); err != nil {
				return err
			}
		}
		return saveWorkbook(book, args[0])
	},
}

var excelSheetsCmd = &cobra.Command{
	Use: "sheets <file>", Short: "List workbook sheets", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		return json.NewEncoder(os.Stdout).Encode(book.GetSheetList())
	},
}

var excelReadCmd = &cobra.Command{
	Use: "read <file> <sheet> [range]", Short: "Read sheet rows as JSON", Args: cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		if len(args) == 3 {
			rows, err := readExcelRange(book, args[1], args[2])
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(rows)
		}
		rows, err := book.GetRows(args[1])
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(rows)
	},
}

func readExcelRange(book *excelize.File, sheet, reference string) ([][]string, error) {
	cells := strings.Split(strings.ToUpper(strings.TrimSpace(reference)), ":")
	if len(cells) == 1 {
		cells = append(cells, cells[0])
	}
	if len(cells) != 2 {
		return nil, fmt.Errorf("invalid range %q; use A1:C3", reference)
	}
	startColumn, startRow, err := excelize.CellNameToCoordinates(cells[0])
	if err != nil {
		return nil, err
	}
	endColumn, endRow, err := excelize.CellNameToCoordinates(cells[1])
	if err != nil {
		return nil, err
	}
	if endColumn < startColumn || endRow < startRow {
		return nil, fmt.Errorf("range end must not precede start")
	}
	rows := make([][]string, 0, endRow-startRow+1)
	for row := startRow; row <= endRow; row++ {
		values := make([]string, 0, endColumn-startColumn+1)
		for column := startColumn; column <= endColumn; column++ {
			cell, err := excelize.CoordinatesToCellName(column, row)
			if err != nil {
				return nil, err
			}
			value, err := book.GetCellValue(sheet, cell)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		rows = append(rows, values)
	}
	return rows, nil
}

var excelGetCmd = &cobra.Command{
	Use: "get <file> <sheet> <cell>", Short: "Read one cell", Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		value, err := book.GetCellValue(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	},
}

var excelSetCmd = &cobra.Command{
	Use: "set <file> <sheet> <cell> <value>", Short: "Set cell; writes --out or --in-place", Args: cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		if err := book.SetCellValue(args[1], args[2], args[3]); err != nil {
			return err
		}
		return saveWorkbookTarget(cmd, book, args[0])
	},
}

var excelAddSheetCmd = &cobra.Command{
	Use: "add-sheet <file> <sheet>", Short: "Add sheet; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		if _, err := book.NewSheet(args[1]); err != nil {
			return err
		}
		return saveWorkbookTarget(cmd, book, args[0])
	},
}

var excelDeleteSheetCmd = &cobra.Command{
	Use: "delete-sheet <file> <sheet>", Short: "Delete sheet; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := excelize.OpenFile(args[0])
		if err != nil {
			return err
		}
		defer book.Close()
		if err := book.DeleteSheet(args[1]); err != nil {
			return err
		}
		return saveWorkbookTarget(cmd, book, args[0])
	},
}

func saveWorkbookTarget(cmd *cobra.Command, book *excelize.File, input string) error {
	out, _ := cmd.Flags().GetString("out")
	inPlace, _ := cmd.Flags().GetBool("in-place")
	if out == "" && !inPlace {
		return fmt.Errorf("choose --out <file> or --in-place")
	}
	if out == "" {
		out = input
	}
	return saveWorkbook(book, out)
}

func saveWorkbook(book *excelize.File, output string) error {
	dir := filepath.Dir(output)
	temp, err := os.CreateTemp(dir, ".unicli-*.xlsx")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempName)
		return err
	}
	defer os.Remove(tempName)
	if err := book.SaveAs(tempName); err != nil {
		return err
	}
	return os.Rename(tempName, output)
}

func init() {
	for _, command := range []*cobra.Command{excelSetCmd, excelAddSheetCmd, excelDeleteSheetCmd} {
		command.Flags().String("out", "", "output workbook path")
		command.Flags().Bool("in-place", false, "replace input workbook")
	}
	excelCmd.AddCommand(excelCreateCmd, excelSheetsCmd, excelReadCmd, excelGetCmd, excelSetCmd, excelAddSheetCmd, excelDeleteSheetCmd)
}
