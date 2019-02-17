// Copyright © 2019 ProgramArkitekten AS leffen@gmail.com
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
	"context"
	"os"
	"strings"

	"github.com/leffen/do-upper/pkg/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logFile            string
	servers            string
	timeBetweenSeconds int64
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a ping server against a given set of web sites",
	Long:  `Checks if a give set of sites is up and timing the response`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// Running with debug info enabled
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(os.Stdout)

		urls := strings.Split(servers, ",")
		s := serve.Server{}
		err := s.Run(ctx, urls, timeBetweenSeconds)
		if err != nil {
			logrus.Fatalf("Unable to serve with error: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&servers, "servers", "s", "", "Urls to check")
	serveCmd.Flags().StringVarP(&logFile, "logfile", "l", "timings.json", "timing log file ( json data) ")
	serveCmd.Flags().Int64VarP(&timeBetweenSeconds, "time-between", "t", 60, "Time beween pings ")
}
