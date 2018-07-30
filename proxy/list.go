// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/requests"
)

// ListFunctions list deployed functions
func ListFunctions(namespace, gateway string, tlsInsecure bool) ([]requests.Function, error) {
	var results []requests.Function

	gateway = strings.TrimRight(gateway, "/")
	timeout := 60 * time.Second
	client := MakeHTTPClient(&timeout, tlsInsecure)

	resourceUrl := ""
	if len(namespace) > 0 {
		resourceUrl = gateway + "/system/functions" + "?namespace=" + namespace
	} else {
		resourceUrl = gateway + "/system/functions"
	}

	fmt.Printf("Making call to list functions with this endpoint: %s\n", resourceUrl)

	getRequest, err := http.NewRequest(http.MethodGet, resourceUrl, nil)
	SetAuth(getRequest, gateway)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	res, err := client.Do(getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", gateway)
		}
		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
		}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return results, nil
}
