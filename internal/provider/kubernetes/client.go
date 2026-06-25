// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vmware/go-vcloud-director/v3/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

const (
	defaultFieldManager = "terraform-provider-vcfa"
)

type Client struct {
	mainClientSet *kubernetes.Clientset
	dynamicClient dynamic.Interface
	warnings      *warningCollector
}

func NewClient(tmClient *vcfa.VCDClient, projectName string, supervisorNamespaceName string) (*Client, error) {
	restConfig, err := getKubernetesRestConfig(tmClient, projectName, supervisorNamespaceName)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes rest config: %w", err)
	}

	restConfig.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return &kubernetesloggingRoundTripper{wrapped: rt}
	}

	warnCollector := &warningCollector{}
	restConfig.WarningHandler = warnCollector

	mainClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes main clientSet: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes dynamic client: %w", err)
	}

	return &Client{mainClientSet: mainClientSet, dynamicClient: dynamicClient, warnings: warnCollector}, nil
}

func (k *Client) ReadClusterScopedResource(ctx context.Context, name string, gvr schema.GroupVersionResource, outType any) error {
	util.Logger.Printf("[K8S] Reading resource %s %s into target type %s", gvr.String(), name, reflect.TypeOf(outType))

	result, err := k.dynamicClient.Resource(gvr).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("error reading resource %s %s: %w", gvr.String(), name, err)
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, outType); err != nil {
		return fmt.Errorf("error converting %s result to resource object %s: %w", gvr.String(), reflect.TypeOf(outType), err)
	}

	return nil
}

func (k *Client) CreateNamespaceScopedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, payload any, outType any, dryRun bool) error {
	util.Logger.Printf("[K8S] Creating resource %s in namespace %s (target type: %s)", gvr.String(), namespace, reflect.TypeOf(outType))

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(payload)
	if err != nil {
		return fmt.Errorf("error converting payload of type %s to unstructured: %w", reflect.TypeOf(payload), err)
	}

	unstructuredCluster := &unstructured.Unstructured{Object: unstructuredMap}
	result, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Create(
		ctx,
		unstructuredCluster,
		metav1.CreateOptions{
			DryRun: func() []string {
				if dryRun {
					return []string{"All"}
				}
				return []string{}
			}(),
			FieldManager: defaultFieldManager,
		},
	)
	if err != nil {
		if dryRun {
			return err
		}
		return fmt.Errorf("error creating resource %s: %w", gvr.String(), err)
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, outType); err != nil {
		return fmt.Errorf("error converting %s result to resource object %s: %w", gvr.String(), reflect.TypeOf(outType), err)
	}

	return nil
}

func (k *Client) ReadNamespaceScopedResource(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource, outType any) error {
	util.Logger.Printf("[K8S] Reading resource %s %s/%s into target type %s", gvr.String(), namespace, name, reflect.TypeOf(outType))

	result, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("error reading resource %s %s/%s: %w", gvr.String(), namespace, name, err)
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, outType); err != nil {
		return fmt.Errorf("error converting %s result to resource object %s: %w", gvr.String(), reflect.TypeOf(outType), err)
	}

	return nil
}

func (k *Client) UpdateNamespaceScopedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, payload any, outType any, dryRun bool) error {
	util.Logger.Printf("[K8S] Updating resource %s in namespace %s (target type: %s)", gvr.String(), namespace, reflect.TypeOf(outType))

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(payload)
	if err != nil {
		return fmt.Errorf("error converting payload of type %s to unstructured: %w", reflect.TypeOf(payload), err)
	}

	unstructuredCluster := &unstructured.Unstructured{Object: unstructuredMap}
	result, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Update(
		ctx,
		unstructuredCluster,
		metav1.UpdateOptions{
			DryRun: func() []string {
				if dryRun {
					return []string{"All"}
				}
				return []string{}
			}(),
			FieldManager: defaultFieldManager,
		},
	)
	if err != nil {
		if dryRun {
			return err
		}
		return fmt.Errorf("error updating resource %s: %w", gvr.String(), err)
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, outType); err != nil {
		return fmt.Errorf("error converting %s result to resource object %s: %w", gvr.String(), reflect.TypeOf(outType), err)
	}

	return nil
}

func (k *Client) PatchNamespaceScopedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, patchType types.PatchType, patchData []byte, outType any, dryRun bool) error {
	util.Logger.Printf("[K8S] Patching (%s)resource %s %s/%s (target type: %s)", gvr.String(), patchType, namespace, name, reflect.TypeOf(outType))

	result, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Patch(
		ctx,
		name,
		patchType,
		patchData,
		metav1.PatchOptions{
			DryRun: func() []string {
				if dryRun {
					return []string{"All"}
				}
				return []string{}
			}(),
			FieldManager: defaultFieldManager,
		},
	)
	if err != nil {
		if dryRun {
			return err
		}
		return fmt.Errorf("error patching (%s) resource %s %s/%s: %w", gvr.String(), patchType, namespace, name, err)
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, outType); err != nil {
		return fmt.Errorf("error converting %s result to resource object %s: %w", gvr.String(), reflect.TypeOf(outType), err)
	}

	return nil
}

func (k *Client) DeleteNamespaceScopedResource(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource, dryRun bool) error {
	util.Logger.Printf("[K8S] Deleting resource %s %s/%s", gvr.String(), namespace, name)

	if err := k.dynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{
			DryRun: func() []string {
				if dryRun {
					return []string{"All"}
				}
				return []string{}
			}(),
		},
	); err != nil {
		if dryRun {
			return err
		}
		return fmt.Errorf("error deleting resource %s %s/%s: %w", gvr.String(), namespace, name, err)
	}

	return nil
}

func (k *Client) ReadSecret(ctx context.Context, namespace string, name string) (*corev1.Secret, error) {
	util.Logger.Printf("[K8S] Reading secret %s/%s", namespace, name)

	secret, err := k.mainClientSet.CoreV1().Secrets(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("error reading secret %s/%s: %w", namespace, name, err)
	}

	return secret, nil
}

func getKubernetesRestConfig(tmClient *vcfa.VCDClient, projectName string, supervisorNamespaceName string) (*rest.Config, error) {
	// Get Supervisor Namespace URL
	clusterName := fmt.Sprintf("%s:%s@%s", tmClient.Org, supervisorNamespaceName, tmClient.Client.VCDHREF.Host)
	contextName := fmt.Sprintf("%s:%s:%s", tmClient.Org, supervisorNamespaceName, projectName)

	supervisorNamespaceEndpointURL, err := helpers.GetSupervisorNamespaceEndpointURL(tmClient, projectName, supervisorNamespaceName)
	if err != nil {
		return nil, err
	}
	clusterServer := supervisorNamespaceEndpointURL

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

	// Build kubeconfig
	kubeconfig := &clientcmdapi.Config{
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
	}

	// Convert to rest.Config using clientcmd API
	configBytes, err := json.Marshal(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error marshaling kubeconfig: %w", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("error building rest config: %w", err)
	}

	return restConfig, nil
}
