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
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/trusch/eve/config"
)

// certaddCmd represents the certadd command
var certaddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a certificate",
	Long: `add a certificate`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		certPath, _ := cmd.Flags().GetString("cert")
		keyPath, _ := cmd.Flags().GetString("key")
		password, _ := cmd.Flags().GetString("password")
		if id==""||certPath==""||keyPath==""||password==""{
			log.Fatal("specify --id, --cert, --key and --password")
		}
		certBs, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Fatal("can not read certificate file")
		}
		keyBs, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatal("can not read key file")
		}
		certConfig := &config.CertConfig{
			ID: id,
			CertPem: string(certBs),
			KeyPem: string(keyBs),
		}
		if err = certConfig.Encrypt(password); err != nil {
			log.Fatal(err)
		}
		if err = client.PutCertConfig(certConfig, true); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	certCmd.AddCommand(certaddCmd)
	certaddCmd.Flags().String("cert", "", "certificate path")
	certaddCmd.Flags().String("key", "", "key path")
	certaddCmd.Flags().String("password", "", "password to encrypt data")
}
