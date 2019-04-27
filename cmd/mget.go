package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	Version    = "0.0.0"
	Buildstamp = time.Now().Format(time.RFC3339)
	Githash    = "custom"

	ProgVersion string
	debugMode   bool

	logger zerolog.Logger

	savePath string
	filelist string
	progress bool
	debug    bool
	version  bool
)

var RootCmd = &cobra.Command{
	Use:              "mget",
	Short:            "mget a command line tool",
	Long:             posixCommandHelp,
	PersistentPreRun: initSettings,
	Run:              rootRun,
}

func init() {
	ProgVersion = fmt.Sprintf("mget-%s", Version)

	// flags
	RootCmd.PersistentFlags().StringVarP(&filelist, "input-file", "i", "", "Read URLs from a local or external file.")
	RootCmd.PersistentFlags().StringVarP(&savePath, "save", "s", "", "give a save path.")
	RootCmd.PersistentFlags().BoolVarP(&progress, "disable-progress", "", false, "disable output download progress information.")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "print debugging messages about its progress.")
	RootCmd.Flags().BoolVarP(&version, "version", "V", false, "output version information and exit")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		logger.Error().Msg(err.Error())
	}
}

// initSettings set log level to debug
func initSettings(cmd *cobra.Command, args []string) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.TimeFieldFormat = time.RFC3339
	logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: true}).With().Timestamp().Logger()
}

func rootRun(cmd *cobra.Command, args []string) {
	if version {
		showVersion()
	}

	var urls []string

	if filelist != "" {
		var err error
		urls, err = readURLsFile(filelist)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
	} else {
		if len(args) < 1 {
			fmt.Println("Please give an URL string!")
			cmd.Help()
		}
		urls = args
	}
	logger.Debug().Msgf("urls: %v", urls)

	if savePath == "" {
		var err error
		savePath, err = os.Getwd()
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
	}

	Downloads(urls, savePath)
}

func VersionString() string {
	return fmt.Sprintf("mget Version %s-%s (%s)", Version, Githash, Buildstamp)

}

func showVersion() {
	fmt.Println(VersionString())
	os.Exit(0)
}
