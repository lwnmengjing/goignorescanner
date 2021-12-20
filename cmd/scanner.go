package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/lwnmengjing/goignorescanner/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scanner",
		Short: "Directory scanner to scan directory and list unignored files",
		Long:  "Directory scanner to scan directory and list unignored files, based on standard ignore files such as .dockerignore, .gitingnore",
		RunE:  scanDir,
	}
	BaseDir    string
	IgnoreType string
)

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVarP(&BaseDir, "basedir", "d", ".", "Directory to scan and list includable files")
	rootCmd.PersistentFlags().StringVarP(&IgnoreType, "ignorefile", "i", ".dockerignore", "The ignore file type to use, the file will be searched for in the basedir root")
}

func scanDir(cmd *cobra.Command, args []string) error {

	ignoreScanner, err := scanner.NewOrDefault(BaseDir, IgnoreType)

	var includes []string
	if ignoreScanner != nil {
		includes, err = ignoreScanner.Scan()
	} else {
		return errors.Unwrap(err)
	}

	if err != nil {
		return errors.Unwrap(err)
	}

	var buf []byte

	buf = append(buf, "\033[22m"...)

	for _, incl := range includes {
		buf = append(buf, "\033[1m"...)
		buf = append(buf, "\n"...)
		buf = append(buf, incl...)
		buf = append(buf, "\033[1m"...)
	}

	_, err = os.Stdout.Write(append(buf, '\n'))

	if err != nil {
		log.Fatal(" os.Stdout.Write =", err)
	}
	return err

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
