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
	"log"
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/trusch/bobbyd/middleware/rule"
)

// mwputCmd represents the mwput command
var mwputCmd = &cobra.Command{
	Use:   "add",
	Short: "add a middleware rule",
	Long: `add a middleware rule`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		route, _ := cmd.Flags().GetString("route")
		middlewareStr, _ := cmd.Flags().GetString("middleware")
		if route == "" || id == "" || middlewareStr == "" {
			log.Fatal("specify --route, --id and --middleware")
		}
		mwRule := &rule.Rule{ID: id, Route: route}
		if err := json.Unmarshal([]byte(middlewareStr), &mwRule.Middlewares); err != nil {
			log.Fatal(err)
		}
		if err := client.PutMwRule(mwRule, true); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	middlewareCmd.AddCommand(mwputCmd)
	mwputCmd.Flags().StringP("route","r","","routing rule")
	mwputCmd.Flags().StringP("middleware","m","","json array describing middleware chain")
}
