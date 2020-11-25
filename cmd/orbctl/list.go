package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/tree"
)

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
			printer.BorderTop, printer.BorderBottom = true, true
			printer.HeaderFgColor = tablewriter.FgYellowColor

			var (
				tail    bool
				headers []string
				rows    [][]string
				cellIdx = -1
			)

			for path, mach := range machines {
				ctx := path[:strings.LastIndex(path, ".")]
				if context == "" || context == ctx {
					v := reflect.ValueOf(mach).Elem()
					if !tail {
						headers = tableprinter.StructParser.ParseHeaders(v)
						if column != "" {
							for idx, h := range headers {
								if strings.Contains(h, column) {
									cellIdx = idx
								}
							}
							if cellIdx == -1 {
								return fmt.Errorf("unknown column: %s", column)
							}
						}
						tail = true
					}

					cells, _ := tableprinter.StructParser.ParseRow(v)

					if cellIdx > -1 {
						fmt.Println(cells[cellIdx])
						continue
					}
					rows = append(rows, cells)
				}
			}

			if cellIdx == -1 {
				printer.Render(headers, rows, nil, false)
			}

			return nil
		})
	}
	return cmd
}
