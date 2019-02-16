// Copyright © 2019 NAME HERE leffen@gmail.com
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

	"github.com/leffen/do-upper/pkg/serve"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a ping server against a given set of web sites",
	Long:  `Checks if a give set of sites is up and timing the response`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		s := serve.Server{}
		s.Run(ctx, []string{"https://leffen.com"}, 10)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

}