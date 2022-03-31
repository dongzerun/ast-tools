package cmd

import (
	"fmt"
	"os"

	"github.com/dongzerun/ast-tools/pkg"
	"github.com/dongzerun/ast-tools/pkg/config"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile = ""
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("mockery")
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if viper.IsSet("config") {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to find homedir")
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(".mockery")
	}

	// Note we purposely ignore the error. Don't care if we can't find a config file.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// NewRootCmd ...
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cmpgen",
		Short: "Generate struct compare functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := GetRootAppFromViper(viper.GetViper())
			if err != nil {
				printStackTrace(err)
				return err
			}
			return r.Run()
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&cfgFile, "config", "", "config file to use")
	pFlags.String("name", "", "name or matching regular expression of interface to generate code")
	pFlags.String("action", "", "action to do, diff, copy, convert")
	pFlags.String("source", "", "source struct, converted to target")
	pFlags.String("src-pkg", "", "src pkg, default is current dir")
	pFlags.String("srctag", "json", "convert field based tag, default is json, can be json, pb")
	pFlags.String("target", "", "target struct")
	pFlags.String("target-pkg", "", "target pkg, default is current dir")
	pFlags.String("targettag", "json", "convert field based tag, default is json, can be json, pb")
	pFlags.Bool("print", false, "print the generated code to stdout")
	pFlags.String("output", "./", "directory to write generated code")
	pFlags.String("outpkg", "", "name of generated package")
	pFlags.String("packageprefix", "", "prefix for the generated package name, it is ignored if outpkg is also specified.")
	pFlags.String("dir", ".", "directory to search for struct")
	pFlags.Bool("inpackage", false, "generate that goes inside the original package")
	pFlags.StringSlice("ignorefields", nil, "ignore fields, like A.b.c,A.F")
	pFlags.String("log-level", "info", "Level of logging")
	pFlags.BoolP("dry-run", "d", false, "Do a dry run, don't modify any files")

	viper.BindPFlags(pFlags)

	cmd.AddCommand(NewShowConfigCmd())
	return cmd
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func printStackTrace(e error) {
	fmt.Printf("%v\n", e)
	if err, ok := e.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			fmt.Printf("%+s:%d\n", f, f)
		}
	}

}

// Execute executes the cobra CLI workflow
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		//printStackTrace(err)
		os.Exit(1)
	}
}

// RootApp2 ...
type RootApp2 struct {
	Ver config.Version
}

// RootApp ...
type RootApp struct {
	Config config.Config
}

// GetRootAppFromViper ...
func GetRootAppFromViper(v *viper.Viper) (*RootApp, error) {
	r := &RootApp{}
	if err := v.UnmarshalExact(&r.Config); err != nil {
		return nil, errors.Wrapf(err, "failed to get config")
	}
	return r, nil
}

// Run ...
func (r *RootApp) Run() error {
	return pkg.NewParser(&r.Config).Do()
}

// NewShowConfigCmd ...
func NewShowConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "showconfig",
		Short: "Show the merged config",
		Long: `Print out a yaml representation of the merged config. 
	This initializes viper and prints out the merged configuration between
	config files, environment variables, and CLI flags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &config.Config{}
			if err := viper.UnmarshalExact(config); err != nil {
				return errors.Wrapf(err, "failed to unmarshal config")
			}
			out, err := yaml.Marshal(config)
			if err != nil {
				return errors.Wrapf(err, "Failed to marsrhal yaml")
			}
			fmt.Printf("%s", string(out))
			return nil
		},
	}
}
