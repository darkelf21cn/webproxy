package main

import (
	"context"
	"gfwpass/conf"
	"gfwpass/service"
	"gfwpass/util"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cmd = &cobra.Command{
	Use:   "passgfw",
	Short: "PassGFW",
	Run: func(cmd *cobra.Command, args []string) {
		runService()
	},
}

var confFile string

func main() {
	pf := cmd.PersistentFlags()
	pf.StringVarP(&confFile, "config", "c", "", "configuration file")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	cmd.Execute()
}

func runService() {
	conf := conf.LoadConfig(viper.GetString("config"))
	logger := util.NewLogger(conf)
	logger.Info("printing config file content", zap.String("conf", conf.String()))
	service, err := service.NewService(context.Background(), logger, conf)
	if err != nil {
		logger.Error("failed to initialize the service", zap.Error(err))
		os.Exit(1)
	}
	service.Start()
}
