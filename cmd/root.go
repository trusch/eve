// Copyright © 2017 Tino Rusch
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trusch/eve/config"
	"github.com/trusch/eve/config/docker"
	"github.com/trusch/eve/config/etcd"
	"github.com/trusch/eve/handler"
	"github.com/trusch/eve/server"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "eve",
	Short: "eve is a virtual entrypoint",
	Long: `eve is a virtual entrypoint.

You can use it to proxy http connections in your container cluster with minimal effort.`,
	Run: func(cmd *cobra.Command, args []string) {
		httpAddr := viper.GetString("http")
		httpsAddr := viper.GetString("https")

		h := handler.New()

		srv, err := server.New(h, httpAddr, httpsAddr)
		if err != nil {
			log.Fatal(err)
		}
		err = srv.ListenAndServeHTTP()
		if err != nil {
			log.Fatal(err)
		}
		err = srv.ListenAndServeHTTPS()
		if err != nil {
			log.Fatal(err)
		}

		configSrcConfigured := false

		etcdAddr := viper.GetString("etcd")
		if etcdAddr != "" {
			cli, err := etcd.NewClient(etcdAddr)
			if err != nil {
				log.Print(err)
			} else {
				configSrcConfigured = true
				go supplyConfig(cli, h, srv)
			}
		}
		if viper.GetBool("docker") {
			cli, err := docker.New()
			if err != nil {
				log.Print(err)
			} else {
				configSrcConfigured = true
				go supplyConfig(cli, h, srv)
			}
		}
		if !configSrcConfigured {
			log.Fatal("specify at least one config source: --docker or --etcd='<etcd-address>'")
		}
		select {}
	},
}

// Execute executes the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.Flags().String("http", ":80", "HTTP Address")
	RootCmd.Flags().String("https", ":443", "HTTPS Address")
	RootCmd.Flags().String("etcd", "127.0.0.1:2379", "etcd server address")
	RootCmd.Flags().Bool("docker", true, "listen for docker events")
	RootCmd.Flags().String("password", "", "certificate seal password")
	RootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.eve.yaml)")
	viper.BindPFlags(RootCmd.Flags())
}

func initConfig() {
	viper.SetConfigName(".eve")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func supplyConfig(src config.Stream, handler *handler.Handler, srv *server.Server) {
	for action := range src.GetChannel() {
		switch action.Type {
		case config.UpsertLbRule:
			{
				log.Print("upsert lb rule: ", action.LbRule)
				handler.LBManager.UpsertRule(action.LbRule)
			}
		case config.UpsertMwRule:
			{
				log.Print("upsert mw rule: ", action.MwRule)
				handler.MWManager.UpsertRule(action.MwRule)
			}
		case config.UpsertHost:
			{
				log.Print("upsert host: ", action.HostConfig)
				handler.LBManager.UpsertServer(action.HostConfig)
			}
		case config.UpsertCert:
			{
				log.Print("upsert cert: ", action.CertConfig.ID)
				cfg := action.CertConfig
				if err := cfg.Decrypt(viper.GetString("password")); err != nil {
					log.Print(err)
					continue
				}
				srv.AddCertificate(cfg.ID, cfg.CertPem, cfg.KeyPem)
				if err := srv.ListenAndServeHTTPS(); err != nil {
					log.Print(err)
				}
			}
		case config.DeleteLbRule:
			{
				log.Print("delete lb rule: ", action.LbRule)
				handler.LBManager.RemoveRule(action.LbRule.ID)
			}
		case config.DeleteMwRule:
			{
				log.Print("delete mw rule: ", action.MwRule)
				handler.MWManager.RemoveRule(action.MwRule.ID)
			}
		case config.DeleteHost:
			{
				log.Print("delete host: ", action.HostConfig)
				handler.LBManager.RemoveServer(action.HostConfig)
			}
		case config.DeleteCert:
			{
				log.Print("delete cert: ", action.CertConfig)
				srv.RemoveCertificate(action.CertConfig.ID)
				if err := srv.ListenAndServeHTTPS(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
