/*
Copyright The Rekor Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sigstore/cosign/pkg/cosign"
)

func SignBlob() *ffcli.Command {
	var (
		flagset = flag.NewFlagSet("cosign sign-blob", flag.ExitOnError)
		key     = flagset.String("key", "", "path to the private key")
		b64     = flagset.Bool("b64", true, "whether to base64 encode the output")
	)
	return &ffcli.Command{
		Name:       "sign-blob",
		ShortUsage: "cosign sign-blob -key <key> [-sig <sig path>] <blob>",
		ShortHelp:  "Sign the supplied blob, outputting the base64-nocded signature to stdout",
		FlagSet:    flagset,
		Exec: func(ctx context.Context, args []string) error {
			if *key == "" {
				return flag.ErrHelp
			}

			if len(args) != 1 {
				return flag.ErrHelp
			}

			return SignBlobCmd(ctx, *key, args[0], *b64, getPass)
		},
	}
}

func SignBlobCmd(ctx context.Context, keyPath, payloadPath string, b64 bool, pf cosign.PassFunc) error {
	var payload []byte
	var err error
	if payloadPath == "-" {
		payload, err = ioutil.ReadAll(os.Stdin)
	} else {
		fmt.Fprintln(os.Stderr, "Using payload from:", payloadPath)
		payload, err = ioutil.ReadFile(payloadPath)
	}
	if err != nil {
		return err
	}

	pass, err := pf(false)
	if err != nil {
		return err
	}
	kb, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}
	pk, err := cosign.LoadPrivateKey(kb, pass)
	if err != nil {
		return err
	}
	signature := ed25519.Sign(pk, payload)
	if b64 {
		fmt.Println(base64.StdEncoding.EncodeToString(signature))
	} else {
		// No newline if using the raw signature
		os.Stdout.Write(signature)
	}
	return nil
}
