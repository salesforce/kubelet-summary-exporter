package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ConfigureTLS(logger *zap.Logger, certAuthorityFile string, insecure bool, nodeHost string) error {
	// Set the root CA pool
	cadata, err := os.ReadFile(certAuthorityFile)
	if err != nil {
		return err
	}
	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(cadata) {
		return fmt.Errorf("failed to append certs from pem")
	}

	newTlsConfig := &tls.Config{}
	newTlsConfig.RootCAs = certs
	if insecure {
		logger.Warn("using insecure tls")
		newTlsConfig.InsecureSkipVerify = insecure
	}

	defaultTransport := http.DefaultTransport.(*http.Transport)
	defaultTransport.TLSClientConfig = newTlsConfig

	dialer := &net.Dialer{
		Timeout:   2 * time.Second,
		KeepAlive: 2 * time.Second,
	}

	defaultTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, fmt.Sprintf("%s:10250", nodeHost))
	}
	return nil
}

// ServerAddrFromCluster uses incluster config to determine a node's Hostname
func ServerAddrFromCluster(nodeHost string) (string, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return "", err
	}

	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeHost, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeHostName {
			return addr.Address, nil
		}
	}
	return "", fmt.Errorf("no node matching %q found", nodeHost)
}
