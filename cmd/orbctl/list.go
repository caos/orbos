package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/tree"
)

type machine struct {
	Context string `header:"context"`
	ID      string `header:"identifier"`
	IP      string `header:"ip address"`
}

func ListCommand(rv RootValues) *cobra.Command {
	var (
		column, context string
		cmd             = &cobra.Command{
			Use:   "list",
			Short: "List available machines",
			Long:  "List available machines",
			Args:  cobra.MaximumNArgs(1),
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&column, "column", "", "Print this column only")
	flags.StringVar(&context, "context", "", "Print machines from this context only")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return machines(monitor, gitClient, orbConfig, func(machineIDs []string, machines map[string]infra.Machine, _ *tree.Tree) error {
			printer := tableprinter.New(os.Stdout)

			var m []machine

			var eachLine func(_ string, mach infra.Machine)
			printTable := false

			switch strings.ToLower(column) {
			case "id":
				fallthrough
			case "identifier":
				eachLine = func(_ string, mach infra.Machine) { fmt.Println(mach.ID()) }
			case "ip":
				fallthrough
			case "addr":
				fallthrough
			case "address":
				fallthrough
			case "ip address":
				eachLine = func(_ string, mach infra.Machine) { fmt.Println(mach.IP()) }
			case "":
				eachLine = func(ctx string, mach infra.Machine) {
					m = append(m, machine{
						Context: ctx,
						ID:      mach.ID(),
						IP:      mach.IP(),
					})
				}
				printTable = true
			default:
				return fmt.Errorf("unknown column: %s", column)
			}

			for path, mach := range machines {
				ctx := path[:strings.LastIndex(path, ".")]
				if context == "" || context == ctx {
					eachLine(ctx, mach)
				}
			}

			if !printTable {
				return nil
			}

			if len(m) == 0 {
				return errors.New("no machines found")
			}

			printer.BorderTop, printer.BorderBottom = true, true
			printer.HeaderFgColor = tablewriter.FgYellowColor
			printer.Print(m)
			return nil
		})
	}
	return cmd
}
