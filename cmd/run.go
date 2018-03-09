// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"sample-extension-apiserver/cmd/server"
	"io"
	//"github.com/golang/glog"
)

func NewCmdRun(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	opt := server.NewOptions(out, errOut)
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch ksd server",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Complete(); err != nil {
				return err
			}
			if err := opt.Validate(args); err != nil {
				return err
			}
			if err := opt.Run(stopCh); err != nil {
				return err
			}

			return nil
		},
	}
	opt.AddFlags(cmd.Flags())

	return cmd
}

//// runCmd represents the run command
//var runCmd = &cobra.Command{
//	Use:   "run",
//	Short: "A brief description of your command",
//	Run: func(cmd *cobra.Command, args []string) {
//		glog.Infoln("starting extension apiserver.........")
//		glog.Infoln("stpCh =", stopCh)
//		glog.Infoln("options =", options)
//		glog.Infoln("etcd =", etcd)
//
//		if err := options.Complete(); err != nil {
//			glog.Fatal(err)
//		}
//		if err := options.Validate(args); err != nil {
//			glog.Fatal(err)
//		}
//		if err := options.Run(stopCh); err != nil {
//			glog.Fatalln("error in starting extension apiserver:", err)
//		}
//	},
//}
//
//func init() {
//	rootCmd.AddCommand(runCmd)
//
//	// Here you will define your flags and configuration settings.
//
//	// Cobra supports Persistent Flags which will work for this command
//	// and all subcommands, e.g.:
//	// runCmd.PersistentFlags().String("foo", "", "A help for foo")
//
//	// Cobra supports local flags which will only run when this command
//	// is called directly, e.g.:
//	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
//}
