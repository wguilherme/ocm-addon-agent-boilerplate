package transmitters

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
)

// configMapTransmitter → cria/atualiza ConfigMap no hub cluster.
// Namespace = SpokeClusterName (criado pelo OCM)
type configMapTransmitter struct {
	configMapName string
}

func NewConfigMapTransmitter(configMapName string) contracts.Transmitter {
	if configMapName == "" {
		configMapName = "cluster-inventory-report"
	}

	return &configMapTransmitter{
		configMapName: configMapName,
	}
}

func (t *configMapTransmitter) Transmit(ctx context.Context, report contracts.ClusterInventoryReport, config *contracts.SyncConfig) error {
	if config == nil {
		return agenterrors.ErrNilConfig
	}
	if config.HubClient == nil {
		return agenterrors.ErrNilHubClient
	}
	if config.SpokeClusterName == "" {
		return agenterrors.ErrEmptyClusterName
	}

	klog.V(4).Infof("[ConfigMapTransmitter] Transmitting report for cluster '%s'", config.SpokeClusterName)

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		klog.Errorf("[ConfigMapTransmitter] Failed to marshal report: %v", err)
		return agenterrors.NewTransmissionError(t.Name(), err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.configMapName,
			Namespace: config.SpokeClusterName,
			Labels: map[string]string{
				"app":     "basic-addon",
				"cluster": config.SpokeClusterName,
				"type":    "inventory-report",
			},
		},
		Data: map[string]string{
			"report": string(reportJSON),
		},
	}

	_, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.V(4).Infof("[ConfigMapTransmitter] ConfigMap already exists, updating...")
			_, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Update(ctx, configMap, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("[ConfigMapTransmitter] Failed to update ConfigMap: %v", err)
				return agenterrors.NewTransmissionError(t.Name(), err)
			}
			klog.V(4).Infof("[ConfigMapTransmitter] ConfigMap updated successfully")
		} else {
			klog.Errorf("[ConfigMapTransmitter] Failed to create ConfigMap: %v", err)
			return agenterrors.NewTransmissionError(t.Name(), err)
		}
	} else {
		klog.V(4).Infof("[ConfigMapTransmitter] ConfigMap created successfully")
	}

	return nil
}

func (t *configMapTransmitter) Name() string {
	return "ConfigMapTransmitter"
}
