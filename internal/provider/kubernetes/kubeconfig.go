// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

func BuildKubeconfig(tmClient *vcfa.VCDClient, clusterServer string, clusterName string, contextName string) (*clientcmdapi.Config, error) {
	// Parse JWT token to extract username.
	// ParseUnverified is intentional: the provider cannot obtain the signing key
	// for VCFA-issued session tokens, so signature verification is not possible.
	// The token is used only to extract the preferred_username claim; it is then
	// forwarded as a bearer token to the Kubernetes API, which performs its own
	// validation against the VCFA identity provider.
	token, _, err := new(jwt.Parser).ParseUnverified(tmClient.Client.VCDToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing JWT token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not parse claims from JWT token")
	}
	preferredUsername, ok := claims["preferred_username"].(string)
	if !ok {
		return nil, errors.New("could not parse preferred username from JWT token claims")
	}
	username := fmt.Sprintf("%s:%s@%s", tmClient.Org, preferredUsername, tmClient.Client.VCDHREF.Host)

	return &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: clientcmdapi.SchemeGroupVersion.Version,
		Clusters: []clientcmdapi.NamedCluster{{
			Name: clusterName,
			Cluster: clientcmdapi.Cluster{
				InsecureSkipTLSVerify: tmClient.InsecureFlag,
				Server:                clusterServer,
			},
		}},
		Contexts: []clientcmdapi.NamedContext{{
			Name: contextName,
			Context: clientcmdapi.Context{
				Cluster:  clusterName,
				AuthInfo: username,
			},
		}},
		AuthInfos: []clientcmdapi.NamedAuthInfo{{
			Name: username,
			AuthInfo: clientcmdapi.AuthInfo{
				Token: token.Raw,
			},
		}},
		CurrentContext: contextName,
	}, nil
}
