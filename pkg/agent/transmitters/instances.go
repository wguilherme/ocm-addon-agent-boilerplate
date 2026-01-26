package transmitters

import "github.com/totvs/addon-framework-basic/pkg/agent/contracts"

var (
	ConfigMapTransmitterInstance contracts.Transmitter
)

func init() {
	ConfigMapTransmitterInstance = NewConfigMapTransmitter("cluster-inventory-report")
}
