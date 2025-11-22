package community

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/fonts"
	"github.com/zachomedia/go-bdf"
)

func NewListFontsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list-fonts",
		Short:             "List available fonts.",
		Example:           `  pixlet community list-fonts`,
		Long:              `This command lists all fonts supported by this Pixlet version.`,
		RunE:              listFontsRun,
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	return cmd
}

type fontEntry struct {
	name       string
	advanceMin int
	advanceMax int
	height     int
	ascent     int
	descent    int
}

func listFontsRun(cmd *cobra.Command, _ []string) error {
	dir, err := fonts.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("could not read fonts: %w", err)
	}

	entries := make([]fontEntry, 0, len(dir))
	for _, entry := range dir {
		if strings.HasSuffix(entry.Name(), fonts.Ext) {
			name := strings.TrimSuffix(entry.Name(), fonts.Ext)

			b, err := fonts.GetBytes(name)
			if err != nil {
				slog.Error("Could not read font", "name", name, "err", err)
				continue
			}

			f, err := bdf.Parse(b)
			if err != nil {
				slog.Error("Could not parse font", "name", name, "err", err)
				continue
			}

			var advanceMin, advanceMax int
			if len(f.Characters) != 0 {
				advanceMin = math.MaxInt
				advanceMax = math.MinInt
				for _, c := range f.Characters {
					advanceMin = min(c.Advance[0], advanceMin)
					advanceMax = max(c.Advance[0], advanceMax)
				}
			}

			entries = append(entries, fontEntry{
				name:       name,
				advanceMin: advanceMin,
				advanceMax: advanceMax,
				height:     f.Ascent + f.Descent,
				ascent:     f.Ascent,
				descent:    f.Descent,
			})
		}
	}

	slices.SortFunc(entries, func(a, b fontEntry) int {
		return cmp.Or(
			cmp.Compare(a.height, b.height),
			cmp.Compare(a.name, b.name),
		)
	})

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	if _, err := w.Write([]byte("NAME\tADVANCE\tHEIGHT\tASCENT\tDESCENT\n")); err != nil {
		return err
	}

	for _, e := range entries {
		advance := strconv.Itoa(e.advanceMin)
		if e.advanceMin != e.advanceMax {
			advance += "-" + strconv.Itoa(e.advanceMax)
		}

		if _, err := fmt.Fprintf(
			w, "%s\t%s\t%d\t%d\t%d\n",
			e.name, advance, e.height, e.ascent, e.descent,
		); err != nil {
			return err
		}
	}

	return w.Flush()
}
