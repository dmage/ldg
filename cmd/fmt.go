package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dmage/ldg/parser"
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [<filename>]",
	Short: "Format a ledger file",
	Run: func(cmd *cobra.Command, args []string) {
		var content []byte
		var err error
		if len(args) == 0 {
			content, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
		} else if len(args) == 1 {
			content, err = ioutil.ReadFile(args[0])
			if err != nil {
				log.Fatal(err)
			}
		} else {
			cmd.Usage()
			os.Exit(1)
		}
		transactions, err := parser.Parse(string(content))
		if err != nil {
			log.Fatal(err)
		}
		for i, tr := range transactions {
			if i != 0 {
				fmt.Printf("\n")
			}
			fmt.Print(tr.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)
}
