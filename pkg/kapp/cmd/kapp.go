package cmd

import (
	"io"

	"github.com/cppforlife/cobrautil"
	"github.com/cppforlife/go-cli-ui/ui"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdac "github.com/k14s/kapp/pkg/kapp/cmd/appchange"
	cmdag "github.com/k14s/kapp/pkg/kapp/cmd/appgroup"
	cmdcm "github.com/k14s/kapp/pkg/kapp/cmd/configmap"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdsa "github.com/k14s/kapp/pkg/kapp/cmd/serviceaccount"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
)

type KappOptions struct {
	ui            *ui.ConfUI
	logger        *logger.UILogger
	configFactory cmdcore.ConfigFactory
	depsFactory   cmdcore.DepsFactory

	UIFlags         UIFlags
	LoggerFlags     LoggerFlags
	KubeconfigFlags cmdcore.KubeconfigFlags
}

func NewKappOptions(ui *ui.ConfUI, configFactory cmdcore.ConfigFactory,
	depsFactory cmdcore.DepsFactory) *KappOptions {

	return &KappOptions{ui: ui, logger: logger.NewUILogger(ui),
		configFactory: configFactory, depsFactory: depsFactory}
}

func NewDefaultKappCmd(ui *ui.ConfUI) *cobra.Command {
	configFactory := cmdcore.NewConfigFactoryImpl()
	depsFactory := cmdcore.NewDepsFactoryImpl(configFactory)
	options := NewKappOptions(ui, configFactory, depsFactory)
	flagsFactory := cmdcore.NewFlagsFactory(configFactory, depsFactory)
	return NewKappCmd(options, flagsFactory)
}

func NewKappCmd(o *KappOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kapp",
		Short: "kapp helps to manage applications on your Kubernetes cluster",

		RunE: cobrautil.ShowHelp,

		// Affects children as well
		SilenceErrors: true,
		SilenceUsage:  true,

		// Disable docs header
		DisableAutoGenTag: true,

		// TODO bash completion
	}

	cmd.SetOutput(uiBlockWriter{o.ui}) // setting output for cmd.Help()

	cmd.SetUsageTemplate(cobrautil.HelpSectionsUsageTemplate([]cobrautil.HelpSection{
		cmdcore.AppHelpGroup,
		cmdcore.AppSupportHelpGroup,
		cmdcore.MiscHelpGroup,
		cmdcore.RestOfCommandsHelpGroup,
	}))

	o.UIFlags.Set(cmd, flagsFactory)
	o.LoggerFlags.Set(cmd, flagsFactory)
	o.KubeconfigFlags.Set(cmd, flagsFactory)

	o.configFactory.ConfigurePathResolver(o.KubeconfigFlags.Path.Value)
	o.configFactory.ConfigureContextResolver(o.KubeconfigFlags.Context.Value)

	cmd.AddCommand(NewVersionCmd(NewVersionOptions(o.ui), flagsFactory))

	cmd.AddCommand(cmdapp.NewListCmd(cmdapp.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewInspectCmd(cmdapp.NewInspectOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeployCmd(cmdapp.NewDeployOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeployConfigCmd(cmdapp.NewDeployConfigOptions(o.ui, o.depsFactory), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeleteCmd(cmdapp.NewDeleteOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewRenameCmd(cmdapp.NewRenameOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewLogsCmd(cmdapp.NewLogsOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewLabelCmd(cmdapp.NewLabelOptions(o.ui, o.depsFactory, o.logger), flagsFactory))

	agCmd := cmdag.NewCmd()
	agCmd.AddCommand(cmdag.NewDeployCmd(cmdag.NewDeployOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	agCmd.AddCommand(cmdag.NewDeleteCmd(cmdag.NewDeleteOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(agCmd)

	cmCmd := cmdcm.NewCmd()
	cmCmd.AddCommand(cmdcm.NewListCmd(cmdcm.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmCmd)

	acCmd := cmdac.NewCmd()
	acCmd.AddCommand(cmdac.NewListCmd(cmdac.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	acCmd.AddCommand(cmdac.NewGCCmd(cmdac.NewGCOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(acCmd)

	saCmd := cmdsa.NewCmd()
	saCmd.AddCommand(cmdsa.NewListCmd(cmdsa.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(saCmd)

	appCmd := cmdtools.NewCmd()
	appCmd.AddCommand(cmdtools.NewInspectCmd(cmdtools.NewInspectOptions(o.ui, o.depsFactory), flagsFactory))
	appCmd.AddCommand(cmdtools.NewDiffCmd(cmdtools.NewDiffOptions(o.ui, o.depsFactory), flagsFactory))
	appCmd.AddCommand(cmdtools.NewListLabelsCmd(cmdtools.NewListLabelsOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(appCmd)

	cmd.AddCommand(NewWebsiteCmd(NewWebsiteOptions()))

	// Last one runs first
	cobrautil.VisitCommands(cmd, func(cmd *cobra.Command) {
		origRunE := cmd.RunE
		if origRunE != nil {
			cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
				defer o.logger.DebugFunc("CommandRun").Finish()
				return origRunE(cmd2, args)
			}
		}
	})

	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureCmdWithSubcmd)
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureLeafCmd)

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(func(*cobra.Command, []string) error {
		o.UIFlags.ConfigureUI(o.ui)
		o.LoggerFlags.Configure(o.logger)
		return nil
	}))

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}

type uiBlockWriter struct {
	ui ui.UI
}

var _ io.Writer = uiBlockWriter{}

func (w uiBlockWriter) Write(p []byte) (n int, err error) {
	w.ui.PrintBlock(p)
	return len(p), nil
}
