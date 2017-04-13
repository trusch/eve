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
	"github.com/trusch/eve/loadbalancer/rule"
	"github.com/spf13/cobra"
)

// lbruleaddCmd represents the lbruleadd command
var lbruleaddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a loadbalancer rule",
	Long: `add a loadbalancer rule`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		target, _ := cmd.Flags().GetString("target")
		route, _ := cmd.Flags().GetString("route")
		if target == "" || id == "" || route == "" {
			log.Fatal("specify --target, --id and --route")
		}
		if err := client.PutLbRule(&rule.Rule{ID: id, Target: target, Route: route}, true); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	ruleCmd.AddCommand(lbruleaddCmd)
	lbruleaddCmd.Flags().StringP("target", "t", "", "target loadbalancer")
	lbruleaddCmd.Flags().StringP("route", "r", "", "routing rule (i.e. Host(\"foo.example.tld\"))")
}
