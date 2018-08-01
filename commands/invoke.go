// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	contentType string
	query       []string
	headers     []string
	invokeAsync bool
	httpMethod  string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)

	invokeCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace should be used")
	invokeCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	invokeCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	invokeCmd.Flags().StringVar(&contentType, "content-type", "text/plain", "The content-type HTTP header such as application/json")
	invokeCmd.Flags().StringArrayVar(&query, "query", []string{}, "pass query-string options")
	invokeCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "pass HTTP request header")
	invokeCmd.Flags().BoolVarP(&invokeAsync, "async", "a", false, "Invoke the function asynchronously")
	invokeCmd.Flags().StringVarP(&httpMethod, "method", "m", "POST", "pass HTTP request method")
	invokeCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	faasCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   `invoke FUNCTION_NAME [--gateway GATEWAY_URL] [--content-type CONTENT_TYPE] [--query PARAM=VALUE] [--header PARAM=VALUE] [--method HTTP_METHOD]`,
	Short: "Invoke an OpenFaaS function",
	Long:  `Invokes an OpenFaaS function and reads from STDIN for the body of the request`,
	Example: `  faas-cli invoke echo --gateway https://domain:port
  faas-cli invoke echo --gateway https://domain:port --content-type application/json
  faas-cli invoke env --query repo=faas-cli --query org=openfaas
  faas-cli invoke env --header X-Ping-Url=http://request.bin/etc
  faas-cli invoke resize-img --async -H "X-Callback-Url=http://gateway:8080/function/send2slack" < image.png
  faas-cli invoke env -H X-Ping-Url=http://request.bin/etc
  faas-cli invoke flask --method GET`,
	RunE: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) error {
	var services stack.Services

	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}
	var yamlGateway string
	functionName = args[0]

	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress := getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read standard input: %s", err.Error())
	}

	response, err := proxy.InvokeFunction(namespace, gatewayAddress, functionName, &functionInput, contentType, query, headers, invokeAsync, httpMethod, tlsInsecure)
	if err != nil {
		return err
	}

	if response != nil {
		os.Stdout.Write(*response)
	}

	return nil
}
