# Arquitetura Analyzer - OCM Addon Agent

## Visão Geral

Arquitetura baseada em **Clean Architecture** e **Use Case Pattern** (inspirado em [totvsapps-platform-service-core](https://dev.azure.com/totvstfs/TOTVSApps-Management/_git/totvsapps-platform-service-core)) para coleta, processamento e transmissão de dados do cluster spoke para o hub.

### Conceitos Principais

1. **Collector[T]**: Coleta recursos Kubernetes do tipo T
2. **Processor[T, R]**: Processa dados T em relatório R
3. **Analyzer[T, R]**: Combina Collector + Processor (composição)
4. **UseCase**: Orquestra múltiplos analyzers em paralelo
5. **Transmitter**: Envia relatórios para o hub
6. **Strategy**: Integra UseCase com addon-framework (SyncStrategy)

---

## Arquitetura em Camadas

```
┌─────────────────────────────────────────────────────────────────┐
│                         SyncStrategy                            │
│                  (Integração com addon-framework)               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                          UseCase                                │
│                   (Orquestração de negócio)                     │
│                                                                 │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │  PodAnalyzer     │  │ ServiceAnalyzer  │  ...               │
│  │                  │  │                  │                    │
│  │  ┌────────────┐  │  │  ┌────────────┐  │                    │
│  │  │ Collector  │  │  │  │ Collector  │  │                    │
│  │  └─────┬──────┘  │  │  └─────┬──────┘  │                    │
│  │        │         │  │        │         │                    │
│  │  ┌─────▼──────┐  │  │  ┌─────▼──────┐  │                    │
│  │  │ Processor  │  │  │  │ Processor  │  │                    │
│  │  └────────────┘  │  │  └────────────┘  │                    │
│  └──────────────────┘  └──────────────────┘                    │
│                                                                 │
│                          ▼                                      │
│                  ┌──────────────┐                               │
│                  │ Transmitter  │                               │
│                  └──────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Estrutura de Pastas

```
pkg/agent/
├── contracts/                  # Interfaces públicas (mockáveis)
│   ├── analyzer.go            # Collector[T], Processor[T,R], Analyzer[T,R], AnalyzerEvents
│   ├── transmitter.go         # Transmitter
│   └── usecase.go             # UseCase[I, O]
│
├── errors/                     # Error hierarchy
│   └── errors.go              # CollectionError, ProcessingError, TransmissionError, ValidationError
│
├── analyzers/                  # Implementações de analyzers
│   ├── pod_collector.go       # PodCollector (Collector[corev1.Pod])
│   ├── pod_processor.go       # PodProcessor (Processor[corev1.Pod, PodAnalysis])
│   ├── pod_analyzer.go        # PodAnalyzer (Analyzer[corev1.Pod, PodAnalysis])
│   ├── pod_analyzer_test.go   # Testes com testify/suite
│   └── instances.go           # Singletons globais
│
├── transmitters/               # Implementações de transmitters
│   ├── configmap_transmitter.go  # ConfigMapTransmitter
│   ├── configmap_transmitter_test.go
│   └── instances.go           # Singletons globais
│
├── usecases/                   # Orquestradores de negócio
│   ├── inventory_usecase.go   # InventoryUseCase (executa analyzers + transmitter)
│   ├── inventory_usecase_test.go
│   └── instances.go           # Singletons globais
│
├── reports/                    # Report types (DTOs)
│   ├── cluster_inventory.go   # ClusterInventoryReport (report final)
│   ├── pod_analysis.go        # PodAnalysis (seção pods)
│   ├── service_analysis.go    # ServiceAnalysis (seção services)
│   ├── ingress_analysis.go    # IngressAnalysis (seção ingresses)
│   └── node_analysis.go       # NodeAnalysis (seção nodes)
│
├── mocks/                      # Mocks genéricos para testes
│   └── mocks.go               # MockCollector[T], MockProcessor[T,R], MockAnalyzer[T,R], etc.
│
└── strategies/                 # Integração com addon-framework
    ├── strategy.go            # Interface SyncStrategy (existente)
    └── inventory.go           # InventoryStrategy (usa InventoryUseCase)
```

---

## Interfaces Core (pkg/agent/contracts/)

### Collector[T]

Coleta recursos brutos do Kubernetes.

```go
type Collector[T any] interface {
    // Collect executa a coleta de dados do spoke cluster
    Collect(ctx context.Context, config *strategies.SyncConfig) ([]T, error)

    // Name retorna identificador para logs/métricas
    Name() string
}
```

**Responsabilidades**:
- Chamadas ao K8s API (List, Get)
- Retornar dados brutos sem processamento
- Validar config (SpokeClient não nil)

---

### Processor[T, R]

Processa dados brutos em relatório estruturado.

```go
type Processor[T any, R any] interface {
    // Process transforma []T em report R
    Process(ctx context.Context, data []T, clusterName string) (R, error)

    // Name retorna identificador para logs/métricas
    Name() string
}
```

**Responsabilidades**:
- Transformar dados brutos em análise
- Calcular métricas, agregações
- Filtrar, validar, enriquecer dados
- **Não fazer chamadas externas** (puro processamento)

---

### Analyzer[T, R]

Combina Collector + Processor com suporte a hooks.

```go
type Analyzer[T any, R any] interface {
    // Analyze executa coleta → processamento
    Analyze(ctx context.Context, config *strategies.SyncConfig) (R, error)

    // Name retorna identificador para logs/métricas
    Name() string

    // WithEvents configura hooks opcionais (fluent interface)
    WithEvents(events *AnalyzerEvents) Analyzer[T, R]
}

// AnalyzerEvents define hooks para injetar comportamento customizado
type AnalyzerEvents struct {
    BeforeCollect  func(ctx context.Context, config *strategies.SyncConfig) error
    AfterCollect   func(ctx context.Context, data any) error
    BeforeProcess  func(ctx context.Context, data any) error
    AfterProcess   func(ctx context.Context, result any) error
}
```

**Responsabilidades**:
- Orquestrar coleta → processamento
- Executar hooks se configurados
- Wrappear erros com contexto (CollectionError, ProcessingError)

---

### UseCase[I, O]

Pattern de Clean Architecture para orquestração de negócio.

```go
type UseCase[I any, O any] interface {
    // Perform executa operação de negócio
    Perform(ctx context.Context, input I) (O, error)
}
```

**Exemplo**: `InventoryUseCase` implementa `UseCase[*SyncConfig, ClusterInventoryReport]`

**Responsabilidades**:
- Executar múltiplos analyzers em paralelo (errgroup)
- Agregar resultados em report final
- Transmitir report para o hub
- **Sem dependência de frameworks** (testável isoladamente)

---

### Transmitter

Envia relatórios para o hub cluster.

```go
type Transmitter interface {
    // Transmit envia report para o hub
    Transmit(ctx context.Context, report reports.ClusterInventoryReport, config *strategies.SyncConfig) error

    // Name retorna identificador para logs/métricas
    Name() string
}
```

**Implementações**:
- `ConfigMapTransmitter`: Cria/atualiza ConfigMap no hub (namespace = spoke cluster name)
- `CRDTransmitter`: Cria/atualiza CRD ClusterInventory no hub (futuro)

---

## Error Hierarchy (pkg/agent/errors/)

Sistema de erros tipados com unwrap support.

```go
type AnalyzerErrorType string

const (
    ErrorTypeCollectionFailed    // Falha na coleta (K8s API)
    ErrorTypeProcessingFailed    // Falha no processamento
    ErrorTypeTransmissionFailed  // Falha no envio ao hub
    ErrorTypeValidationFailed    // Falha na validação
)

// Erros wrappam causa original (errors.Unwrap suportado)
type CollectionError struct {
    AnalyzerName string
    Cause        error
}

type ProcessingError struct {
    ProcessorName string
    Cause         error
}

type TransmissionError struct {
    TransmitterName string
    Cause           error
}

// Erros pré-definidos
var (
    ErrNilConfig         = &ValidationError{Field: "config", Message: "..."}
    ErrNilSpokeClient    = &ValidationError{Field: "config.SpokeClient", Message: "..."}
    ErrEmptyClusterName  = &ValidationError{Field: "config.SpokeClusterName", Message: "..."}
)
```

**Uso**:
```go
if err != nil {
    return agenterrors.NewCollectionError("PodAnalyzer", err)
}
```

---

## Report Types (pkg/agent/reports/)

### ClusterInventoryReport

Report final agregado com todas as análises.

```go
type ClusterInventoryReport struct {
    ClusterName string    `json:"clusterName"`
    Timestamp   time.Time `json:"timestamp"`

    // Seções preenchidas por analyzers
    PodAnalysis     *PodAnalysis     `json:"podAnalysis,omitempty"`
    ServiceAnalysis *ServiceAnalysis `json:"serviceAnalysis,omitempty"`
    IngressAnalysis *IngressAnalysis `json:"ingressAnalysis,omitempty"`
    NodeAnalysis    *NodeAnalysis    `json:"nodeAnalysis,omitempty"`
}
```

### PodAnalysis

```go
type PodAnalysis struct {
    TotalPods   int            `json:"totalPods"`
    RunningPods int            `json:"runningPods"`
    PendingPods int            `json:"pendingPods"`
    FailedPods  int            `json:"failedPods"`
    PodsByPhase map[string]int `json:"podsByPhase"`
    Pods        []PodInfo      `json:"pods"`
}

type PodInfo struct {
    Name      string `json:"name"`
    Namespace string `json:"namespace"`
    Phase     string `json:"phase"`
    NodeName  string `json:"nodeName,omitempty"`
}
```

---

## Implementação Completa: PodAnalyzer

### 1. PodCollector

```go
// pkg/agent/analyzers/pod_collector.go
type podCollector struct {
    namespaces []string // Se vazio, coleta de todos
}

func NewPodCollector() contracts.Collector[corev1.Pod] {
    return &podCollector{namespaces: []string{}}
}

func (c *podCollector) Collect(ctx context.Context, config *strategies.SyncConfig) ([]corev1.Pod, error) {
    if config == nil {
        return nil, agenterrors.ErrNilConfig
    }
    if config.SpokeClient == nil {
        return nil, agenterrors.ErrNilSpokeClient
    }

    podList, err := config.SpokeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    return podList.Items, nil
}

func (c *podCollector) Name() string {
    return "PodCollector"
}
```

---

### 2. PodProcessor

```go
// pkg/agent/analyzers/pod_processor.go
type podProcessor struct{}

func NewPodProcessor() contracts.Processor[corev1.Pod, reports.PodAnalysis] {
    return &podProcessor{}
}

func (p *podProcessor) Process(ctx context.Context, pods []corev1.Pod, clusterName string) (reports.PodAnalysis, error) {
    analysis := reports.PodAnalysis{
        TotalPods:   len(pods),
        PodsByPhase: make(map[string]int),
        Pods:        make([]reports.PodInfo, 0, len(pods)),
    }

    for _, pod := range pods {
        phase := string(pod.Status.Phase)
        analysis.PodsByPhase[phase]++

        switch pod.Status.Phase {
        case corev1.PodRunning:
            analysis.RunningPods++
        case corev1.PodPending:
            analysis.PendingPods++
        case corev1.PodFailed:
            analysis.FailedPods++
        }

        analysis.Pods = append(analysis.Pods, reports.PodInfo{
            Name:      pod.Name,
            Namespace: pod.Namespace,
            Phase:     phase,
            NodeName:  pod.Spec.NodeName,
        })
    }

    return analysis, nil
}

func (p *podProcessor) Name() string {
    return "PodProcessor"
}
```

---

### 3. PodAnalyzer (Composição + Hooks)

```go
// pkg/agent/analyzers/pod_analyzer.go
type podAnalyzer struct {
    collector contracts.Collector[corev1.Pod]
    processor contracts.Processor[corev1.Pod, reports.PodAnalysis]
    events    *contracts.AnalyzerEvents
}

func NewPodAnalyzer() contracts.Analyzer[corev1.Pod, reports.PodAnalysis] {
    return &podAnalyzer{
        collector: NewPodCollector(),
        processor: NewPodProcessor(),
        events:    nil,
    }
}

// Para testes - injeção de dependências
func NewPodAnalyzerWithDeps(
    collector contracts.Collector[corev1.Pod],
    processor contracts.Processor[corev1.Pod, reports.PodAnalysis],
) contracts.Analyzer[corev1.Pod, reports.PodAnalysis] {
    if collector == nil {
        panic("collector cannot be nil")
    }
    if processor == nil {
        panic("processor cannot be nil")
    }
    return &podAnalyzer{collector: collector, processor: processor, events: nil}
}

func (a *podAnalyzer) Analyze(ctx context.Context, config *strategies.SyncConfig) (reports.PodAnalysis, error) {
    // Hook: BeforeCollect
    if a.events != nil && a.events.BeforeCollect != nil {
        if err := a.events.BeforeCollect(ctx, config); err != nil {
            klog.Warningf("[PodAnalyzer] BeforeCollect hook failed: %v", err)
        }
    }

    // 1. Coleta
    pods, err := a.collector.Collect(ctx, config)
    if err != nil {
        return reports.PodAnalysis{}, agenterrors.NewCollectionError(a.Name(), err)
    }

    // Hook: AfterCollect
    if a.events != nil && a.events.AfterCollect != nil {
        if err := a.events.AfterCollect(ctx, pods); err != nil {
            klog.Warningf("[PodAnalyzer] AfterCollect hook failed: %v", err)
        }
    }

    // Hook: BeforeProcess
    if a.events != nil && a.events.BeforeProcess != nil {
        if err := a.events.BeforeProcess(ctx, pods); err != nil {
            klog.Warningf("[PodAnalyzer] BeforeProcess hook failed: %v", err)
        }
    }

    // 2. Processamento
    analysis, err := a.processor.Process(ctx, pods, config.SpokeClusterName)
    if err != nil {
        return reports.PodAnalysis{}, agenterrors.NewProcessingError(a.processor.Name(), err)
    }

    // Hook: AfterProcess
    if a.events != nil && a.events.AfterProcess != nil {
        if err := a.events.AfterProcess(ctx, analysis); err != nil {
            klog.Warningf("[PodAnalyzer] AfterProcess hook failed: %v", err)
        }
    }

    return analysis, nil
}

func (a *podAnalyzer) Name() string {
    return "PodAnalyzer"
}

func (a *podAnalyzer) WithEvents(events *contracts.AnalyzerEvents) contracts.Analyzer[corev1.Pod, reports.PodAnalysis] {
    return &podAnalyzer{
        collector: a.collector,
        processor: a.processor,
        events:    events,
    }
}
```

---

### 4. Singletons (Dependency Injection Manual)

```go
// pkg/agent/analyzers/instances.go
var (
    PodAnalyzerInstance contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
)

func init() {
    PodAnalyzerInstance = NewPodAnalyzer()
}
```

---

## InventoryUseCase (Orquestrador)

Executa múltiplos analyzers em paralelo e transmite resultado.

```go
// pkg/agent/usecases/inventory_usecase.go
type InventoryUseCase struct {
    podAnalyzer contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
    // serviceAnalyzer contracts.Analyzer[corev1.Service, reports.ServiceAnalysis]
    // ingressAnalyzer ...
    transmitter contracts.Transmitter
}

func NewInventoryUseCase(
    podAnalyzer contracts.Analyzer[corev1.Pod, reports.PodAnalysis],
    transmitter contracts.Transmitter,
) contracts.UseCase[*strategies.SyncConfig, reports.ClusterInventoryReport] {
    if podAnalyzer == nil {
        panic("podAnalyzer cannot be nil")
    }
    if transmitter == nil {
        panic("transmitter cannot be nil")
    }
    return &InventoryUseCase{
        podAnalyzer: podAnalyzer,
        transmitter: transmitter,
    }
}

func (u *InventoryUseCase) Perform(ctx context.Context, config *strategies.SyncConfig) (reports.ClusterInventoryReport, error) {
    // Validações
    if config == nil {
        return reports.ClusterInventoryReport{}, agenterrors.ErrNilConfig
    }
    if config.SpokeClusterName == "" {
        return reports.ClusterInventoryReport{}, agenterrors.ErrEmptyClusterName
    }

    report := reports.ClusterInventoryReport{
        ClusterName: config.SpokeClusterName,
        Timestamp:   time.Now().UTC(),
    }

    // Mutex para escrita concorrente no report
    var mu sync.Mutex

    // Executar analyzers em paralelo
    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        podAnalysis, err := u.podAnalyzer.Analyze(ctx, config)
        if err != nil {
            return err
        }
        mu.Lock()
        report.PodAnalysis = &podAnalysis
        mu.Unlock()
        return nil
    })

    // TODO: Adicionar outros analyzers aqui

    // Aguardar todos os analyzers
    if err := g.Wait(); err != nil {
        return reports.ClusterInventoryReport{}, err
    }

    // Transmitir
    if err := u.transmitter.Transmit(ctx, report, config); err != nil {
        return reports.ClusterInventoryReport{}, err
    }

    return report, nil
}
```

**Singleton**:

```go
// pkg/agent/usecases/instances.go
var (
    InventoryUseCaseInstance contracts.UseCase[*strategies.SyncConfig, reports.ClusterInventoryReport]
)

func init() {
    InventoryUseCaseInstance = NewInventoryUseCase(
        analyzers.PodAnalyzerInstance,
        transmitters.ConfigMapTransmitterInstance,
    )
}
```

---

## InventoryStrategy (Integração com addon-framework)

```go
// pkg/agent/strategies/inventory.go
type InventoryStrategy struct {
    useCase contracts.UseCase[*SyncConfig, reports.ClusterInventoryReport]
}

func NewInventoryStrategy() SyncStrategy {
    return &InventoryStrategy{
        useCase: usecases.InventoryUseCaseInstance,
    }
}

func (s *InventoryStrategy) Sync(ctx context.Context, config *SyncConfig) error {
    _, err := s.useCase.Perform(ctx, config)
    return err
}
```

**Registrar no agent.go**:

```go
syncStrategies := []strategies.SyncStrategy{
    strategies.NewHelloStrategy(),
    strategies.NewInventoryStrategy(), // NOVO
}
```

---

## ConfigMapTransmitter

```go
// pkg/agent/transmitters/configmap_transmitter.go
type configMapTransmitter struct {
    configMapName string
}

func NewConfigMapTransmitter(configMapName string) contracts.Transmitter {
    if configMapName == "" {
        configMapName = "cluster-inventory-report"
    }
    return &configMapTransmitter{configMapName: configMapName}
}

func (t *configMapTransmitter) Transmit(ctx context.Context, report reports.ClusterInventoryReport, config *strategies.SyncConfig) error {
    if config == nil {
        return agenterrors.ErrNilConfig
    }
    if config.HubClient == nil {
        return agenterrors.ErrNilHubClient
    }

    reportJSON, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
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

    // Create or Update
    _, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Create(ctx, configMap, metav1.CreateOptions{})
    if err != nil {
        if apierrors.IsAlreadyExists(err) {
            _, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Update(ctx, configMap, metav1.UpdateOptions{})
        }
        if err != nil {
            return agenterrors.NewTransmissionError(t.Name(), err)
        }
    }

    return nil
}

func (t *configMapTransmitter) Name() string {
    return "ConfigMapTransmitter"
}
```

---

## Testes (testify/suite)

### Test Suite para PodAnalyzer

```go
// pkg/agent/analyzers/pod_analyzer_test.go
type PodAnalyzerTestSuite struct {
    suite.Suite
    mockCollector *mocks.MockCollector[corev1.Pod]
    mockProcessor *mocks.MockProcessor[corev1.Pod, reports.PodAnalysis]
    analyzer      contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
    config        *strategies.SyncConfig
}

func (s *PodAnalyzerTestSuite) SetupTest() {
    s.mockCollector = new(mocks.MockCollector[corev1.Pod])
    s.mockProcessor = new(mocks.MockProcessor[corev1.Pod, reports.PodAnalysis])
    s.analyzer = NewPodAnalyzerWithDeps(s.mockCollector, s.mockProcessor)
    s.config = &strategies.SyncConfig{SpokeClusterName: "test-cluster"}
}

func (s *PodAnalyzerTestSuite) TestCollectionError() {
    expectedErr := errors.New("k8s api error")
    s.mockCollector.On("Collect", mock.Anything, s.config).Return(nil, expectedErr)

    result, err := s.analyzer.Analyze(context.Background(), s.config)

    s.Error(err)
    s.Empty(result)

    var collectionErr *agenterrors.CollectionError
    s.True(errors.As(err, &collectionErr))
    s.Equal("PodAnalyzer", collectionErr.AnalyzerName)

    s.mockCollector.AssertExpectations(s.T())
    s.mockProcessor.AssertNotCalled(s.T(), "Process")
}

func (s *PodAnalyzerTestSuite) TestSuccessfulAnalysis() {
    pods := []corev1.Pod{{Status: corev1.PodStatus{Phase: corev1.PodRunning}}}
    analysis := reports.PodAnalysis{TotalPods: 1, RunningPods: 1}

    s.mockCollector.On("Collect", mock.Anything, s.config).Return(pods, nil)
    s.mockProcessor.On("Process", mock.Anything, pods, "test-cluster").Return(analysis, nil)

    result, err := s.analyzer.Analyze(context.Background(), s.config)

    s.NoError(err)
    s.Equal(1, result.TotalPods)
    s.mockCollector.AssertExpectations(s.T())
    s.mockProcessor.AssertExpectations(s.T())
}

func TestPodAnalyzerTestSuite(t *testing.T) {
    suite.Run(t, new(PodAnalyzerTestSuite))
}
```

---

## Como Adicionar Novo Analyzer

### 1. Criar Report Type

```go
// pkg/agent/reports/deployment_analysis.go
type DeploymentAnalysis struct {
    TotalDeployments int              `json:"totalDeployments"`
    ReadyDeployments int              `json:"readyDeployments"`
    Deployments      []DeploymentInfo `json:"deployments"`
}

type DeploymentInfo struct {
    Name      string `json:"name"`
    Namespace string `json:"namespace"`
    Replicas  int32  `json:"replicas"`
    Ready     int32  `json:"ready"`
}
```

### 2. Criar Collector, Processor, Analyzer

Seguir mesmo pattern de `pod_collector.go`, `pod_processor.go`, `pod_analyzer.go`.

### 3. Adicionar no InventoryUseCase

```go
type InventoryUseCase struct {
    podAnalyzer        contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
    deploymentAnalyzer contracts.Analyzer[appsv1.Deployment, reports.DeploymentAnalysis] // NOVO
    transmitter        contracts.Transmitter
}

// No Perform(), adicionar:
g.Go(func() error {
    deploymentAnalysis, err := u.deploymentAnalyzer.Analyze(ctx, config)
    if err != nil {
        return err
    }
    mu.Lock()
    report.DeploymentAnalysis = &deploymentAnalysis
    mu.Unlock()
    return nil
})
```

### 4. Adicionar campo no ClusterInventoryReport

```go
type ClusterInventoryReport struct {
    // ...
    DeploymentAnalysis *DeploymentAnalysis `json:"deploymentAnalysis,omitempty"` // NOVO
}
```

---

## Benefícios da Arquitetura

| Benefício | Descrição |
|-----------|-----------|
| **Clean Architecture** | UseCase separa orquestração de infraestrutura |
| **Dependency Injection** | Singletons + `NewXWithDeps()` para testes |
| **Error Hierarchy** | Erros tipados facilitam tratamento específico |
| **Hooks/Events** | Callbacks opcionais para métricas, logs, transformações |
| **Test Suite Pattern** | testify/suite com setup/teardown automático |
| **Type Safety** | Generics garantem type safety em compile-time |
| **Paralelismo** | errgroup executa analyzers concorrentemente |
| **Extensibilidade** | Adicionar analyzer = 3 arquivos (collector, processor, analyzer) |
| **Manutenibilidade** | Cada camada é independente e testável isoladamente |

---

## Melhorias do service-core Aplicadas

1. ✅ **Contracts separados** (pkg/agent/contracts/)
2. ✅ **Error hierarchy** (pkg/agent/errors/)
3. ✅ **UseCase pattern** (pkg/agent/usecases/)
4. ✅ **Dependency injection manual** (singletons em instances.go)
5. ✅ **Hooks/Events pattern** (AnalyzerEvents)
6. ✅ **Test suite pattern** (testify/suite)
7. ✅ **Fail-fast validations** (panic em construtores se deps nil)

---

## Próximos Passos

### Implementação
- [x] Contracts (interfaces públicas)
- [x] Error hierarchy
- [x] PodAnalyzer completo (Collector + Processor + Analyzer)
- [x] ConfigMapTransmitter
- [x] InventoryUseCase
- [x] InventoryStrategy
- [x] Testes com testify/suite
- [ ] ServiceAnalyzer, IngressAnalyzer, NodeAnalyzer
- [ ] CRDTransmitter (quando CRD estiver definida)

### Melhorias Futuras
- [ ] Retry logic no Transmitter
- [ ] Métricas Prometheus por analyzer
- [ ] Cache de dados coletados
- [ ] Configuração via YAML (quais analyzers habilitar)
- [ ] Rate limiting
- [ ] Worker pattern para scheduling

---

## Referências

- [OCM Addon Framework](https://open-cluster-management.io/concepts/addon/)
- [totvsapps-platform-service-core](https://dev.azure.com/totvstfs/TOTVSApps-Management/_git/totvsapps-platform-service-core)
- [Go Generics](https://go.dev/doc/tutorial/generics)
- [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [testify/suite](https://pkg.go.dev/github.com/stretchr/testify/suite)
