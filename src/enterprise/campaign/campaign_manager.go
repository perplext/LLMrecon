package campaign

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
)

// CampaignManager manages attack campaigns
type CampaignManager struct {
    mu              sync.RWMutex
    campaigns       map[string]*Campaign
    orchestrator    *AttackOrchestrator
    scheduler       *CampaignScheduler
    monitor         *CampaignMonitor
    analytics       *CampaignAnalytics
    playbooks       map[string]*Playbook
    config          ManagerConfig
}

// ManagerConfig holds configuration for campaign manager
type ManagerConfig struct {
    MaxConcurrentCampaigns int
    MaxAttacksPerCampaign  int
    AutomationLevel        AutomationLevel
    MonitoringInterval     time.Duration
    RetentionPeriod        time.Duration
}

// AutomationLevel defines campaign automation level
type AutomationLevel string

const (
    AutomationManual       AutomationLevel = "manual"
    AutomationSemiAuto     AutomationLevel = "semi_auto"
    AutomationFullAuto     AutomationLevel = "full_auto"
    AutomationAdaptive     AutomationLevel = "adaptive"
)

// Campaign represents an attack campaign
type Campaign struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Objectives      []Objective            `json:"objectives"`
    Targets         []*Target              `json:"targets"`
    Phases          []*Phase               `json:"phases"`
    Playbook        *Playbook              `json:"playbook"`
    Schedule        *Schedule              `json:"schedule"`
    Resources       *ResourceAllocation    `json:"resources"`
    Status          CampaignStatus         `json:"status"`
    Progress        *CampaignProgress      `json:"progress"`
    Results         *CampaignResults       `json:"results"`
    Metadata        map[string]interface{} `json:"metadata"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}

// Objective represents a campaign objective
type Objective struct {
    ID          string              `json:"id"`
    Title       string              `json:"title"`
    Description string              `json:"description"`
    Priority    Priority            `json:"priority"`
    Criteria    []SuccessCriterion  `json:"criteria"`
    Status      ObjectiveStatus     `json:"status"`
}

// Priority defines objective priority
type Priority string

const (
    PriorityCritical Priority = "critical"
    PriorityHigh     Priority = "high"
    PriorityMedium   Priority = "medium"
    PriorityLow      Priority = "low"
)

// SuccessCriterion defines success criteria
type SuccessCriterion struct {
    Metric    string      `json:"metric"`
    Operator  string      `json:"operator"`
    Value     interface{} `json:"value"`
    Achieved  bool        `json:"achieved"`
}

// ObjectiveStatus defines objective status
type ObjectiveStatus string

const (
    ObjectivePending   ObjectiveStatus = "pending"
    ObjectiveActive    ObjectiveStatus = "active"
    ObjectiveCompleted ObjectiveStatus = "completed"
    ObjectiveFailed    ObjectiveStatus = "failed"
)

// Target represents an attack target
type Target struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Type         TargetType             `json:"type"`
    Endpoint     string                 `json:"endpoint"`
    Model        string                 `json:"model"`
    Version      string                 `json:"version"`
    AccessLevel  string                 `json:"access_level"`
    Constraints  []string               `json:"constraints"`
    Profile      *TargetProfile         `json:"profile"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// TargetType defines target types
type TargetType string

const (
    TargetChatbot       TargetType = "chatbot"
    TargetAPI           TargetType = "api"
    TargetCodeAssistant TargetType = "code_assistant"
    TargetContentGen    TargetType = "content_generator"
    TargetCustom        TargetType = "custom"
)

// TargetProfile contains target profiling data
type TargetProfile struct {
    Capabilities    []string               `json:"capabilities"`
    Guardrails      []string               `json:"guardrails"`
    KnownWeaknesses []string               `json:"known_weaknesses"`
    ResponsePatterns map[string]interface{} `json:"response_patterns"`
    LastProfiled    time.Time             `json:"last_profiled"`
}

// Phase represents a campaign phase
type Phase struct {
    ID              string          `json:"id"`
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    Order           int             `json:"order"`
    AttackSequences []*AttackSequence `json:"attack_sequences"`
    Prerequisites   []string        `json:"prerequisites"`
    Duration        time.Duration   `json:"duration"`
    Status          PhaseStatus     `json:"status"`
    StartedAt       *time.Time      `json:"started_at,omitempty"`
    CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}

// PhaseStatus defines phase status
type PhaseStatus string

const (
    PhasePending    PhaseStatus = "pending"
    PhaseActive     PhaseStatus = "active"
    PhaseCompleted  PhaseStatus = "completed"
    PhaseFailed     PhaseStatus = "failed"
    PhaseSkipped    PhaseStatus = "skipped"
)

// AttackSequence represents a sequence of attacks
type AttackSequence struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Attacks     []*AttackSpec  `json:"attacks"`
    Flow        FlowType       `json:"flow"`
    Conditions  []Condition    `json:"conditions"`
    MaxRetries  int            `json:"max_retries"`
    Timeout     time.Duration  `json:"timeout"`
}

// AttackSpec specifies an attack
type AttackSpec struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Technique  string                 `json:"technique"`
    Payload    string                 `json:"payload"`
    Parameters map[string]interface{} `json:"parameters"`
    Target     string                 `json:"target"`
    Success    *SuccessMetrics        `json:"success,omitempty"`
}

// FlowType defines attack flow types
type FlowType string

const (
    FlowSequential FlowType = "sequential"
    FlowParallel   FlowType = "parallel"
    FlowConditional FlowType = "conditional"
    FlowIterative  FlowType = "iterative"
)

// Condition represents an execution condition
type Condition struct {
    Type     ConditionType `json:"type"`
    Check    string        `json:"check"`
    Value    interface{}   `json:"value"`
    Action   string        `json:"action"`
}

// ConditionType defines condition types
type ConditionType string

const (
    ConditionSuccess    ConditionType = "success"
    ConditionFailure    ConditionType = "failure"
    ConditionThreshold  ConditionType = "threshold"
)

// SuccessMetrics defines success metrics
type SuccessMetrics struct {
    Achieved      bool                   `json:"achieved"`
    Score         float64                `json:"score"`
    Metrics       map[string]interface{} `json:"metrics"`
    Evidence      []string               `json:"evidence"`
}

// Playbook represents an attack playbook
type Playbook struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Version     string                 `json:"version"`
    Author      string                 `json:"author"`
    Tags        []string               `json:"tags"`
    Tactics     []*Tactic              `json:"tactics"`
    Techniques  map[string]*Technique  `json:"techniques"`
    Procedures  map[string]*Procedure  `json:"procedures"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Tactic represents an attack tactic
type Tactic struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Techniques  []string   `json:"techniques"`
    Order       int        `json:"order"`
}

// Technique represents an attack technique
type Technique struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    Procedures  []string               `json:"procedures"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Procedure represents an attack procedure
type Procedure struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Steps        []Step                 `json:"steps"`
    Requirements []string               `json:"requirements"`
    Outputs      []string               `json:"outputs"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// Step represents a procedure step
type Step struct {
    Order       int                    `json:"order"`
    Action      string                 `json:"action"`
    Parameters  map[string]interface{} `json:"parameters"`
    Validation  string                 `json:"validation"`
}

// Schedule represents campaign schedule
type Schedule struct {
    Type           ScheduleType   `json:"type"`
    StartTime      time.Time      `json:"start_time"`
    EndTime        *time.Time     `json:"end_time,omitempty"`
    Intervals      []Interval     `json:"intervals,omitempty"`
    TimeZone       string         `json:"timezone"`
    Blackouts      []Blackout     `json:"blackouts,omitempty"`
}

// ScheduleType defines schedule types
type ScheduleType string

const (
    ScheduleImmediate  ScheduleType = "immediate"
    ScheduleDelayed    ScheduleType = "delayed"
    ScheduleRecurring  ScheduleType = "recurring"
    ScheduleTriggered  ScheduleType = "triggered"
)

// Interval represents a time interval
type Interval struct {
    Start    time.Time     `json:"start"`
    End      time.Time     `json:"end"`
    Repeat   RepeatPattern `json:"repeat,omitempty"`
}

// RepeatPattern defines repeat patterns
type RepeatPattern struct {
    Frequency string `json:"frequency"`
    Count     int    `json:"count"`
}

// Blackout represents a blackout period
type Blackout struct {
    Start  time.Time `json:"start"`
    End    time.Time `json:"end"`
    Reason string    `json:"reason"`
}

// ResourceAllocation represents resource allocation
type ResourceAllocation struct {
    Compute      ComputeResources       `json:"compute"`
    Network      NetworkResources       `json:"network"`
    Storage      StorageResources       `json:"storage"`
    Concurrency  int                    `json:"concurrency"`
    RateLimits   map[string]RateLimit   `json:"rate_limits"`
}

// ComputeResources defines compute resources
type ComputeResources struct {
    CPU    int `json:"cpu"`
    Memory int `json:"memory"`
    GPU    int `json:"gpu,omitempty"`
}

// NetworkResources defines network resources
type NetworkResources struct {
    Bandwidth   int      `json:"bandwidth"`
    Connections int      `json:"connections"`
    Proxies     []string `json:"proxies,omitempty"`
}

// StorageResources defines storage resources
type StorageResources struct {
    Size      int    `json:"size"`
    Type      string `json:"type"`
    Retention string `json:"retention"`
}

// RateLimit defines rate limits
type RateLimit struct {
    Requests int           `json:"requests"`
    Window   time.Duration `json:"window"`
}

// CampaignStatus defines campaign status
type CampaignStatus string

const (
    CampaignDraft      CampaignStatus = "draft"
    CampaignScheduled  CampaignStatus = "scheduled"
    CampaignRunning    CampaignStatus = "running"
    CampaignPaused     CampaignStatus = "paused"
    CampaignCompleted  CampaignStatus = "completed"
    CampaignFailed     CampaignStatus = "failed"
    CampaignAborted    CampaignStatus = "aborted"
)

// CampaignProgress tracks campaign progress
type CampaignProgress struct {
    CurrentPhase      string                 `json:"current_phase"`
    CompletedPhases   []string               `json:"completed_phases"`
    TotalAttacks      int                    `json:"total_attacks"`
    ExecutedAttacks   int                    `json:"executed_attacks"`
    SuccessfulAttacks int                    `json:"successful_attacks"`
    FailedAttacks     int                    `json:"failed_attacks"`
    Percentage        float64                `json:"percentage"`
    EstimatedTime     time.Duration          `json:"estimated_time"`
    Metrics           map[string]interface{} `json:"metrics"`
}

// CampaignResults contains campaign results
type CampaignResults struct {
    Summary         ResultSummary          `json:"summary"`
    Findings        []*Finding             `json:"findings"`
    Vulnerabilities []*Vulnerability       `json:"vulnerabilities"`
    Statistics      map[string]interface{} `json:"statistics"`
    Timeline        []*TimelineEvent       `json:"timeline"`
    Artifacts       []*Artifact            `json:"artifacts"`
}

// ResultSummary summarizes results
type ResultSummary struct {
    ObjectivesAchieved int                    `json:"objectives_achieved"`
    TotalObjectives    int                    `json:"total_objectives"`
    OverallScore       float64                `json:"overall_score"`
    Risk               RiskLevel              `json:"risk"`
    Recommendations    []string               `json:"recommendations"`
    KeyFindings        []string               `json:"key_findings"`
}

// RiskLevel defines risk levels
type RiskLevel string

const (
    RiskCritical RiskLevel = "critical"
    RiskHigh     RiskLevel = "high"
    RiskMedium   RiskLevel = "medium"
    RiskLow      RiskLevel = "low"
)

// Finding represents a campaign finding
type Finding struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Severity    string                 `json:"severity"`
    Impact      string                 `json:"impact"`
    Evidence    []Evidence             `json:"evidence"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Vulnerability represents a discovered vulnerability
type Vulnerability struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Category     string                 `json:"category"`
    Severity     string                 `json:"severity"`
    Exploitable  bool                   `json:"exploitable"`
    Remediation  string                 `json:"remediation"`
    References   []string               `json:"references"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// Evidence represents attack evidence
type Evidence struct {
    Type      string    `json:"type"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
}

// TimelineEvent represents a timeline event
type TimelineEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Phase       string                 `json:"phase"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Artifact represents a campaign artifact
type Artifact struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Type     string    `json:"type"`
    Path     string    `json:"path"`
    Size     int64     `json:"size"`
    Created  time.Time `json:"created"`
    Metadata map[string]interface{} `json:"metadata"`
}

// NewCampaignManager creates a new campaign manager
func NewCampaignManager(config ManagerConfig) *CampaignManager {
    return &CampaignManager{
        campaigns:    make(map[string]*Campaign),
        orchestrator: NewAttackOrchestrator(),
        scheduler:    NewCampaignScheduler(),
        monitor:      NewCampaignMonitor(config.MonitoringInterval),
        analytics:    NewCampaignAnalytics(),
        playbooks:    make(map[string]*Playbook),
        config:       config,
    }
}

// CreateCampaign creates a new campaign
func (cm *CampaignManager) CreateCampaign(ctx context.Context, campaign *Campaign) (*Campaign, error) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if len(cm.campaigns) >= cm.config.MaxConcurrentCampaigns {
        return nil, fmt.Errorf("maximum concurrent campaigns limit reached")
    }

    campaign.ID = generateCampaignID()
    campaign.Status = CampaignDraft
    campaign.CreatedAt = time.Now()
    campaign.UpdatedAt = time.Now()

    // Initialize progress tracking
    campaign.Progress = &CampaignProgress{
        CurrentPhase:    "",
        CompletedPhases: []string{},
        Metrics:         make(map[string]interface{}),
    }

    // Initialize results
    campaign.Results = &CampaignResults{
        Findings:        []*Finding{},
        Vulnerabilities: []*Vulnerability{},
        Statistics:      make(map[string]interface{}),
        Timeline:        []*TimelineEvent{},
        Artifacts:       []*Artifact{},
    }

    // Validate campaign structure
    if err := cm.validateCampaign(campaign); err != nil {
        return nil, fmt.Errorf("campaign validation failed: %w", err)
    }

    cm.campaigns[campaign.ID] = campaign

    // Load playbook if specified
    if campaign.Playbook != nil && campaign.Playbook.ID != "" {
        if err := cm.loadPlaybook(campaign.Playbook.ID); err != nil {
            return nil, fmt.Errorf("failed to load playbook: %w", err)
        }
    }

    return campaign, nil
}

// StartCampaign starts a campaign
func (cm *CampaignManager) StartCampaign(ctx context.Context, campaignID string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    campaign, exists := cm.campaigns[campaignID]
    if !exists {
        return fmt.Errorf("campaign not found")
    }

    if campaign.Status != CampaignDraft && campaign.Status != CampaignScheduled {
        return fmt.Errorf("campaign cannot be started from status: %s", campaign.Status)
    }

    // Schedule campaign execution
    if err := cm.scheduler.Schedule(campaign); err != nil {
        return fmt.Errorf("failed to schedule campaign: %w", err)
    }

    campaign.Status = CampaignRunning
    campaign.UpdatedAt = time.Now()

    // Start monitoring
    cm.monitor.StartMonitoring(campaign)

    // Begin orchestration
    go cm.orchestrator.Execute(ctx, campaign)

    return nil
}

// AttackOrchestrator orchestrates attack execution
type AttackOrchestrator struct {
    executors map[string]AttackExecutor
    mu        sync.RWMutex
}

// AttackExecutor interface for attack execution
type AttackExecutor interface {
    Execute(ctx context.Context, spec *AttackSpec) (*SuccessMetrics, error)
}

// NewAttackOrchestrator creates a new attack orchestrator
func NewAttackOrchestrator() *AttackOrchestrator {
    return &AttackOrchestrator{
        executors: make(map[string]AttackExecutor),
    }
}

// Execute executes a campaign
func (ao *AttackOrchestrator) Execute(ctx context.Context, campaign *Campaign) error {
    // Execute phases in order
    for _, phase := range campaign.Phases {
        if err := ao.executePhase(ctx, phase, campaign); err != nil {
            campaign.Status = CampaignFailed
            return fmt.Errorf("phase %s failed: %w", phase.Name, err)
        }
    }

    campaign.Status = CampaignCompleted
    return nil
}

// executePhase executes a campaign phase
func (ao *AttackOrchestrator) executePhase(ctx context.Context, phase *Phase, campaign *Campaign) error {
    phase.Status = PhaseActive
    startTime := time.Now()
    phase.StartedAt = &startTime

    // Update campaign progress
    campaign.Progress.CurrentPhase = phase.ID

    // Execute attack sequences
    for _, sequence := range phase.AttackSequences {
        if err := ao.executeSequence(ctx, sequence, campaign); err != nil {
            phase.Status = PhaseFailed
            return err
        }
    }

    phase.Status = PhaseCompleted
    completedTime := time.Now()
    phase.CompletedAt = &completedTime
    
    campaign.Progress.CompletedPhases = append(campaign.Progress.CompletedPhases, phase.ID)
    
    return nil
}

// executeSequence executes an attack sequence
func (ao *AttackOrchestrator) executeSequence(ctx context.Context, sequence *AttackSequence, campaign *Campaign) error {
    switch sequence.Flow {
    case FlowSequential:
        return ao.executeSequential(ctx, sequence, campaign)
    case FlowParallel:
        return ao.executeParallel(ctx, sequence, campaign)
    case FlowConditional:
        return ao.executeConditional(ctx, sequence, campaign)
    case FlowIterative:
        return ao.executeIterative(ctx, sequence, campaign)
    default:
        return fmt.Errorf("unknown flow type: %s", sequence.Flow)
    }
}

// executeSequential executes attacks sequentially
func (ao *AttackOrchestrator) executeSequential(ctx context.Context, sequence *AttackSequence, campaign *Campaign) error {
    for _, attack := range sequence.Attacks {
        if err := ao.executeAttack(ctx, attack, campaign); err != nil {
            return err
        }
    }
    return nil
}

// executeParallel executes attacks in parallel
func (ao *AttackOrchestrator) executeParallel(ctx context.Context, sequence *AttackSequence, campaign *Campaign) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(sequence.Attacks))

    for _, attack := range sequence.Attacks {
        wg.Add(1)
        go func(a *AttackSpec) {
            defer wg.Done()
            if err := ao.executeAttack(ctx, a, campaign); err != nil {
                errChan <- err
            }
        }(attack)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}

// executeConditional executes attacks based on conditions
func (ao *AttackOrchestrator) executeConditional(ctx context.Context, sequence *AttackSequence, campaign *Campaign) error {
    for _, attack := range sequence.Attacks {
        // Check conditions before execution
        if ao.checkConditions(sequence.Conditions, campaign) {
            if err := ao.executeAttack(ctx, attack, campaign); err != nil {
                return err
            }
        }
    }
    return nil
}

// executeIterative executes attacks iteratively
func (ao *AttackOrchestrator) executeIterative(ctx context.Context, sequence *AttackSequence, campaign *Campaign) error {
    for i := 0; i < sequence.MaxRetries; i++ {
        success := true
        for _, attack := range sequence.Attacks {
            if err := ao.executeAttack(ctx, attack, campaign); err != nil {
                success = false
                break
            }
        }
        if success {
            return nil
        }
    }
    return fmt.Errorf("iterative sequence failed after %d retries", sequence.MaxRetries)
}

// executeAttack executes a single attack
func (ao *AttackOrchestrator) executeAttack(ctx context.Context, attack *AttackSpec, campaign *Campaign) error {
    campaign.Progress.TotalAttacks++
    
    executor, exists := ao.executors[attack.Type]
    if !exists {
        return fmt.Errorf("no executor for attack type: %s", attack.Type)
    }

    metrics, err := executor.Execute(ctx, attack)
    if err != nil {
        campaign.Progress.FailedAttacks++
        return err
    }

    attack.Success = metrics
    campaign.Progress.ExecutedAttacks++
    if metrics.Achieved {
        campaign.Progress.SuccessfulAttacks++
    }

    // Update progress percentage
    campaign.Progress.Percentage = float64(campaign.Progress.ExecutedAttacks) / float64(campaign.Progress.TotalAttacks) * 100

    return nil
}

// checkConditions checks execution conditions
func (ao *AttackOrchestrator) checkConditions(conditions []Condition, campaign *Campaign) bool {
    for _, condition := range conditions {
        // Implement condition checking logic
        switch condition.Type {
        case ConditionSuccess:
            // Check if previous attacks succeeded
        case ConditionFailure:
            // Check if previous attacks failed
        case ConditionThreshold:
            // Check if metrics meet threshold
        case ConditionTime:
            // Check time-based conditions
        }
    }
    return true
}

// CampaignScheduler handles campaign scheduling
type CampaignScheduler struct {
    schedules map[string]*Schedule
    mu        sync.RWMutex
}

// NewCampaignScheduler creates a new campaign scheduler
func NewCampaignScheduler() *CampaignScheduler {
    return &CampaignScheduler{
        schedules: make(map[string]*Schedule),
    }
}

// Schedule schedules a campaign
func (cs *CampaignScheduler) Schedule(campaign *Campaign) error {
    cs.mu.Lock()
    defer cs.mu.Unlock()

    if campaign.Schedule == nil {
        campaign.Schedule = &Schedule{
            Type:      ScheduleImmediate,
            StartTime: time.Now(),
        }
    }

    cs.schedules[campaign.ID] = campaign.Schedule
    return nil
}

// CampaignMonitor monitors campaign execution
type CampaignMonitor struct {
    interval time.Duration
    monitors map[string]chan bool
    mu       sync.RWMutex
}

// NewCampaignMonitor creates a new campaign monitor
func NewCampaignMonitor(interval time.Duration) *CampaignMonitor {
    return &CampaignMonitor{
        interval: interval,
        monitors: make(map[string]chan bool),
    }
}

// StartMonitoring starts monitoring a campaign
func (cm *CampaignMonitor) StartMonitoring(campaign *Campaign) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    stopChan := make(chan bool)
    cm.monitors[campaign.ID] = stopChan

    go cm.monitorCampaign(campaign, stopChan)
}

// monitorCampaign monitors a single campaign
func (cm *CampaignMonitor) monitorCampaign(campaign *Campaign, stop chan bool) {
    ticker := time.NewTicker(cm.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Collect metrics and update progress
            cm.updateMetrics(campaign)
        case <-stop:
            return
        }
    }
}

// updateMetrics updates campaign metrics
func (cm *CampaignMonitor) updateMetrics(campaign *Campaign) {
    // Implement metric collection
    campaign.Progress.Metrics["last_update"] = time.Now()
}

// CampaignAnalytics provides campaign analytics
type CampaignAnalytics struct {
    mu sync.RWMutex
}

// NewCampaignAnalytics creates new campaign analytics
func NewCampaignAnalytics() *CampaignAnalytics {
    return &CampaignAnalytics{}
}

// Helper functions
func generateCampaignID() string {
    return fmt.Sprintf("campaign_%d", time.Now().UnixNano())
}

// validateCampaign validates campaign structure
func (cm *CampaignManager) validateCampaign(campaign *Campaign) error {
    if len(campaign.Objectives) == 0 {
        return fmt.Errorf("campaign must have at least one objective")
    }

    if len(campaign.Targets) == 0 {
        return fmt.Errorf("campaign must have at least one target")
    }

    if len(campaign.Phases) == 0 {
        return fmt.Errorf("campaign must have at least one phase")
    }

    totalAttacks := 0
    for _, phase := range campaign.Phases {
        for _, sequence := range phase.AttackSequences {
            totalAttacks += len(sequence.Attacks)
        }
    }

    if totalAttacks > cm.config.MaxAttacksPerCampaign {
        return fmt.Errorf("campaign exceeds maximum attacks limit: %d > %d", 
            totalAttacks, cm.config.MaxAttacksPerCampaign)
    }

    return nil
}

// loadPlaybook loads a playbook
func (cm *CampaignManager) loadPlaybook(playbookID string) error {
    // Implement playbook loading logic
    return nil
}

// PauseCampaign pauses a running campaign
func (cm *CampaignManager) PauseCampaign(ctx context.Context, campaignID string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    campaign, exists := cm.campaigns[campaignID]
    if !exists {
        return fmt.Errorf("campaign not found")
    }

    if campaign.Status != CampaignRunning {
        return fmt.Errorf("can only pause running campaigns")
    }

    campaign.Status = CampaignPaused
    campaign.UpdatedAt = time.Now()

    return nil
}

// ResumeCampaign resumes a paused campaign
func (cm *CampaignManager) ResumeCampaign(ctx context.Context, campaignID string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    campaign, exists := cm.campaigns[campaignID]
    if !exists {
        return fmt.Errorf("campaign not found")
    }

    if campaign.Status != CampaignPaused {
        return fmt.Errorf("can only resume paused campaigns")
    }

    campaign.Status = CampaignRunning
    campaign.UpdatedAt = time.Now()

    // Resume orchestration
    go cm.orchestrator.Execute(ctx, campaign)

    return nil
}

// GetCampaignStatus gets campaign status and progress
func (cm *CampaignManager) GetCampaignStatus(campaignID string) (*Campaign, error) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    campaign, exists := cm.campaigns[campaignID]
    if !exists {
        return nil, fmt.Errorf("campaign not found")
    }

    return campaign, nil
}