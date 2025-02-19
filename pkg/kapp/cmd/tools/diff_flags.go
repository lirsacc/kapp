package tools

import (
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type DiffFlags struct {
	ctlcap.ChangeSetViewOpts
	ctldiff.ChangeSetOpts

	Run bool
}

func (s *DiffFlags) SetWithPrefix(prefix string, cmd *cobra.Command) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.Run, prefix+"run", false, "Show diff and exit successfully without any further action")

	cmd.Flags().BoolVar(&s.Summary, prefix+"summary", true, "Show diff summary")
	cmd.Flags().BoolVarP(&s.Changes, prefix+"changes", "c", false, "Show changes")

	cmd.Flags().IntVar(&s.Context, prefix+"context", 2, "Show number of lines around changed lines")
	cmd.Flags().BoolVar(&s.AgainstLastApplied, prefix+"against-last-applied", true, "Show changes against last applied copy when possible")
}
