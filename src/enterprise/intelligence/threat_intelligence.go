package intelligence

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
)

// ThreatIntelligenceSystem manages threat intelligence for LLM attacks
type ThreatIntelligenceSystem struct {
    mu              sync.RWMutex
    feeds           map[string]*ThreatFeed
    indicators      map[string]*ThreatIndicator
    vulnerabilities map[string]*VulnerabilityIntel
    actors          map[string]*ThreatActor
    campaigns       map[string]*ThreatCampaign
    analyzer        *IntelligenceAnalyzer
    correlator      *ThreatCorrelator
    predictor       *ThreatPredictor
    repository      *IntelRepository
    config          IntelConfig
}

}
// IntelConfig holds configuration for threat intelligence
type IntelConfig struct {
    MaxFeeds           int
    UpdateInterval     time.Duration
    RetentionPeriod    time.Duration
    AnalysisDepth      int
    AutoCorrelation    bool
    PredictionEnabled  bool

}
// ThreatFeed represents a threat intelligence feed
type ThreatFeed struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Type        FeedType            `json:"type"`
    Source      string              `json:"source"`
    URL         string              `json:"url"`
    Format      string              `json:"format"`
    Frequency   time.Duration       `json:"frequency"`
    LastUpdate  time.Time          `json:"last_update"`
    Status      FeedStatus          `json:"status"`
    Reliability ReliabilityScore    `json:"reliability"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// FeedType defines types of threat feeds
type FeedType string

const (
    FeedVulnerability   FeedType = "vulnerability"
    FeedIndicator       FeedType = "indicator"
    FeedTactic          FeedType = "tactic"
    FeedActor           FeedType = "actor"
    FeedIncident        FeedType = "incident"
    FeedResearch        FeedType = "research"
)

// FeedStatus defines feed status
type FeedStatus string

const (
    FeedActive    FeedStatus = "active"
    FeedInactive  FeedStatus = "inactive"
    FeedError     FeedStatus = "error"
    FeedUpdating  FeedStatus = "updating"
)

// ReliabilityScore defines feed reliability
type ReliabilityScore float64

const (
    ReliabilityUnknown     ReliabilityScore = 0.0
    ReliabilityLow         ReliabilityScore = 0.25
    ReliabilityMedium      ReliabilityScore = 0.50
    ReliabilityHigh        ReliabilityScore = 0.75
    ReliabilityVeryHigh    ReliabilityScore = 1.0
)

// ThreatIndicator represents a threat indicator
type ThreatIndicator struct {
    ID              string                 `json:"id"`
    Type            IndicatorType          `json:"type"`
    Value           string                 `json:"value"`
    Pattern         string                 `json:"pattern"`
    Description     string                 `json:"description"`
    Severity        SeverityLevel          `json:"severity"`
    Confidence      ConfidenceLevel        `json:"confidence"`
    FirstSeen       time.Time             `json:"first_seen"`
    LastSeen        time.Time             `json:"last_seen"`
    ValidUntil      *time.Time            `json:"valid_until,omitempty"`
    Tags            []string              `json:"tags"`
    RelatedIOCs     []string              `json:"related_iocs"`
    Mitigations     []Mitigation          `json:"mitigations"`
    Metadata        map[string]interface{} `json:"metadata"`

}
// IndicatorType defines indicator types
type IndicatorType string

const (
    IndicatorPromptPattern    IndicatorType = "prompt_pattern"
    IndicatorPayload          IndicatorType = "payload"
    IndicatorBehavior         IndicatorType = "behavior"
    IndicatorExfiltration     IndicatorType = "exfiltration"
    IndicatorEvasion          IndicatorType = "evasion"
    IndicatorPersistence      IndicatorType = "persistence"
)

// SeverityLevel defines severity levels
type SeverityLevel string

const (
    SeverityCritical    SeverityLevel = "critical"
    SeverityHigh        SeverityLevel = "high"
    SeverityMedium      SeverityLevel = "medium"
    SeverityLow         SeverityLevel = "low"
    SeverityInfo        SeverityLevel = "info"
)

// ConfidenceLevel defines confidence levels
type ConfidenceLevel string

const (
    ConfidenceConfirmed     ConfidenceLevel = "confirmed"
    ConfidenceHigh          ConfidenceLevel = "high"
    ConfidenceMedium        ConfidenceLevel = "medium"
    ConfidenceLow           ConfidenceLevel = "low"
    ConfidenceUnknown       ConfidenceLevel = "unknown"
)

// Mitigation represents a mitigation strategy
type Mitigation struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Type        MitigationType         `json:"type"`
    Effectiveness float64              `json:"effectiveness"`
    Implementation string               `json:"implementation"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// MitigationType defines mitigation types
type MitigationType string

const (
    MitigationPreventive    MitigationType = "preventive"
    MitigationDetective     MitigationType = "detective"
    MitigationCorrective    MitigationType = "corrective"
    MitigationCompensating  MitigationType = "compensating"
)

// VulnerabilityIntel represents vulnerability intelligence
type VulnerabilityIntel struct {
    ID              string                 `json:"id"`
    CVE             string                 `json:"cve,omitempty"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Category        VulnCategory           `json:"category"`
    Severity        SeverityLevel          `json:"severity"`
    CVSS            *CVSSScore             `json:"cvss,omitempty"`
    AffectedModels  []string              `json:"affected_models"`
    ExploitAvailable bool                  `json:"exploit_available"`
    ExploitCode     string                 `json:"exploit_code,omitempty"`
    Patches         []Patch                `json:"patches"`
    Workarounds     []string              `json:"workarounds"`
    References      []string              `json:"references"`
    DiscoveredDate  time.Time             `json:"discovered_date"`
    PublishedDate   time.Time             `json:"published_date"`
    LastModified    time.Time             `json:"last_modified"`
    Metadata        map[string]interface{} `json:"metadata"`
}

}
// VulnCategory defines vulnerability categories
type VulnCategory string

const (
    VulnPromptInjection     VulnCategory = "prompt_injection"
    VulnJailbreak           VulnCategory = "jailbreak"
    VulnDataLeakage         VulnCategory = "data_leakage"
    VulnModelExtraction     VulnCategory = "model_extraction"
    VulnSupplyChain         VulnCategory = "supply_chain"
    VulnAccessControl       VulnCategory = "access_control"
    VulnDenialOfService     VulnCategory = "denial_of_service"
)

// CVSSScore represents CVSS scoring
type CVSSScore struct {
    Version         string  `json:"version"`
    BaseScore       float64 `json:"base_score"`
    TemporalScore   float64 `json:"temporal_score"`
    EnvironmentalScore float64 `json:"environmental_score"`
    Vector          string  `json:"vector"`
}

}
// Patch represents a security patch
type Patch struct {
    ID              string    `json:"id"`
    Version         string    `json:"version"`
    ReleaseDate     time.Time `json:"release_date"`
    Description     string    `json:"description"`
    DownloadURL     string    `json:"download_url"`
}

}
// ThreatActor represents a threat actor
type ThreatActor struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Aliases         []string              `json:"aliases"`
    Type            ActorType              `json:"type"`
    Motivation      []string              `json:"motivation"`
    Sophistication  SophisticationLevel    `json:"sophistication"`
    Resources       ResourceLevel          `json:"resources"`
    Intent          []string              `json:"intent"`
    Capabilities    []string              `json:"capabilities"`
    TTPs            []TTP                  `json:"ttps"`
    TargetSectors   []string              `json:"target_sectors"`
    OriginCountry   string                 `json:"origin_country,omitempty"`
    Active          bool                   `json:"active"`
    FirstSeen       time.Time             `json:"first_seen"`
    LastSeen        time.Time             `json:"last_seen"`
    Metadata        map[string]interface{} `json:"metadata"`

}
// ActorType defines threat actor types
type ActorType string

const (
    ActorNationState    ActorType = "nation_state"
    ActorCriminal       ActorType = "criminal"
    ActorHacktivist     ActorType = "hacktivist"
    ActorInsider        ActorType = "insider"
    ActorResearcher     ActorType = "researcher"
    ActorUnknown        ActorType = "unknown"
)

// SophisticationLevel defines sophistication levels
type SophisticationLevel string

const (
    SophisticationNovice        SophisticationLevel = "novice"
    SophisticationIntermediate  SophisticationLevel = "intermediate"
    SophisticationAdvanced      SophisticationLevel = "advanced"
    SophisticationExpert        SophisticationLevel = "expert"
)

// ResourceLevel defines resource levels
type ResourceLevel string

const (
    ResourceIndividual      ResourceLevel = "individual"
    ResourceGroup           ResourceLevel = "group"
    ResourceOrganization    ResourceLevel = "organization"
    ResourceGovernment      ResourceLevel = "government"
)

// TTP represents Tactics, Techniques, and Procedures
type TTP struct {
    ID          string                 `json:"id"`
    Tactic      string                 `json:"tactic"`
    Technique   string                 `json:"technique"`
    Procedure   string                 `json:"procedure"`
    Description string                 `json:"description"`
    Examples    []string              `json:"examples"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// ThreatCampaign represents a threat campaign
type ThreatCampaign struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Actors          []string              `json:"actors"`
    Targets         []CampaignTarget      `json:"targets"`
    Timeline        CampaignTimeline      `json:"timeline"`
    Objectives      []string              `json:"objectives"`
    TTPs            []TTP                 `json:"ttps"`
    Indicators      []string              `json:"indicators"`
    Impact          ImpactAssessment      `json:"impact"`
    Status          CampaignStatus        `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`

}
// CampaignTarget represents a campaign target
type CampaignTarget struct {
    Type        string   `json:"type"`
    Sector      string   `json:"sector"`
    Geography   []string `json:"geography"`
    Systems     []string `json:"systems"`

}
// CampaignTimeline represents campaign timeline
type CampaignTimeline struct {
    Started     time.Time  `json:"started"`
    Ended       *time.Time `json:"ended,omitempty"`
    Active      bool       `json:"active"`
    Duration    string     `json:"duration"`

}
// ImpactAssessment represents impact assessment
type ImpactAssessment struct {
    Severity        SeverityLevel          `json:"severity"`
    Scope           string                 `json:"scope"`
    DataCompromised bool                   `json:"data_compromised"`
    SystemsAffected int                    `json:"systems_affected"`
    FinancialImpact string                 `json:"financial_impact,omitempty"`
    Reputation      string                 `json:"reputation_impact,omitempty"`

}
// CampaignStatus defines campaign status
type CampaignStatus string

const (
    CampaignActive      CampaignStatus = "active"
    CampaignDormant     CampaignStatus = "dormant"
    CampaignCompleted   CampaignStatus = "completed"
    CampaignSuspected   CampaignStatus = "suspected"
)

// IntelligenceAnalyzer analyzes threat intelligence
type IntelligenceAnalyzer struct {
    patterns    map[string]*Pattern
    trends      map[string]*Trend
    predictions map[string]*Prediction
    mu          sync.RWMutex

}
// Pattern represents an identified pattern
type Pattern struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Frequency   int                    `json:"frequency"`
    Indicators  []string              `json:"indicators"`
    FirstSeen   time.Time             `json:"first_seen"`
    LastSeen    time.Time             `json:"last_seen"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// Trend represents a threat trend
type Trend struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        TrendType              `json:"type"`
    Direction   TrendDirection         `json:"direction"`
    Magnitude   float64                `json:"magnitude"`
    TimeWindow  time.Duration          `json:"time_window"`
    DataPoints  []DataPoint            `json:"data_points"`
    Forecast    []ForecastPoint        `json:"forecast"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// TrendType defines trend types
type TrendType string

const (
    TrendAttackVolume       TrendType = "attack_volume"
    TrendTechniqueEvolution TrendType = "technique_evolution"
    TrendActorActivity      TrendType = "actor_activity"
    TrendVulnerability      TrendType = "vulnerability"
)

// TrendDirection defines trend directions
type TrendDirection string

const (
    TrendIncreasing TrendDirection = "increasing"
    TrendDecreasing TrendDirection = "decreasing"
    TrendStable     TrendDirection = "stable"
    TrendVolatile   TrendDirection = "volatile"
)

// DataPoint represents a data point
type DataPoint struct {
    Timestamp time.Time   `json:"timestamp"`
    Value     float64     `json:"value"`
    Label     string      `json:"label"`
    Metadata  interface{} `json:"metadata,omitempty"`

}
// ForecastPoint represents a forecast point
type ForecastPoint struct {
    Timestamp   time.Time `json:"timestamp"`
    Value       float64   `json:"value"`
    Confidence  float64   `json:"confidence"`
    Upper       float64   `json:"upper_bound"`
    Lower       float64   `json:"lower_bound"`
}

}
// Prediction represents a threat prediction
type Prediction struct {
    ID              string                 `json:"id"`
    Type            PredictionType         `json:"type"`
    Description     string                 `json:"description"`
    Probability     float64                `json:"probability"`
    Impact          SeverityLevel          `json:"impact"`
    TimeFrame       string                 `json:"timeframe"`
    Indicators      []string              `json:"indicators"`
    Recommendations []string              `json:"recommendations"`
    Confidence      ConfidenceLevel        `json:"confidence"`
    CreatedAt       time.Time             `json:"created_at"`
    ValidUntil      time.Time             `json:"valid_until"`
    Metadata        map[string]interface{} `json:"metadata"`

}
// PredictionType defines prediction types
type PredictionType string

const (
    PredictionAttack        PredictionType = "attack"
    PredictionVulnerability PredictionType = "vulnerability"
    PredictionCampaign      PredictionType = "campaign"
    PredictionTechnique     PredictionType = "technique"
)

// ThreatCorrelator correlates threat data
type ThreatCorrelator struct {
    correlations map[string]*Correlation
    rules        map[string]*CorrelationRule
    mu           sync.RWMutex
}

}
// Correlation represents a threat correlation
type Correlation struct {
    ID          string                 `json:"id"`
    Type        CorrelationType        `json:"type"`
    Entities    []CorrelatedEntity     `json:"entities"`
    Strength    float64                `json:"strength"`
    Evidence    []string              `json:"evidence"`
    CreatedAt   time.Time             `json:"created_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// CorrelationType defines correlation types
type CorrelationType string

const (
    CorrelationActorIOC         CorrelationType = "actor_ioc"
    CorrelationCampaignTTP      CorrelationType = "campaign_ttp"
    CorrelationVulnExploit      CorrelationType = "vuln_exploit"
    CorrelationIncidentPattern  CorrelationType = "incident_pattern"
)

// CorrelatedEntity represents a correlated entity
type CorrelatedEntity struct {
    Type    string `json:"type"`
    ID      string `json:"id"`
    Role    string `json:"role"`
    Weight  float64 `json:"weight"`

}
// CorrelationRule defines correlation rules
type CorrelationRule struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Conditions  []RuleCondition        `json:"conditions"`
    Actions     []RuleAction           `json:"actions"`
    Priority    int                    `json:"priority"`
    Enabled     bool                   `json:"enabled"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// RuleCondition represents a rule condition
type RuleCondition struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
}

}
// RuleAction represents a rule action
type RuleAction struct {
    Type       string                 `json:"type"`
    Parameters map[string]interface{} `json:"parameters"`

}
// ThreatPredictor predicts threats
type ThreatPredictor struct {
    models      map[string]*PredictionModel
    historical  *HistoricalData
    mu          sync.RWMutex
}

}
// PredictionModel represents a prediction model
type PredictionModel struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Version     string                 `json:"version"`
    Accuracy    float64                `json:"accuracy"`
    Features    []string              `json:"features"`
    LastTrained time.Time             `json:"last_trained"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// HistoricalData represents historical threat data
type HistoricalData struct {
    TimeRange   TimeRange              `json:"time_range"`
    Incidents   []HistoricalIncident   `json:"incidents"`
    Statistics  map[string]interface{} `json:"statistics"`
}

}
// TimeRange represents a time range
type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`

}
// HistoricalIncident represents a historical incident
type HistoricalIncident struct {
    ID          string                 `json:"id"`
    Date        time.Time             `json:"date"`
    Type        string                 `json:"type"`
    Severity    SeverityLevel          `json:"severity"`
    Actors      []string              `json:"actors"`
    Techniques  []string              `json:"techniques"`
    Impact      string                 `json:"impact"`
    Resolution  string                 `json:"resolution"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// IntelRepository manages intelligence storage
type IntelRepository struct {
    storage     map[string]interface{}
    indices     map[string]map[string]string
    mu          sync.RWMutex
}

}
// NewThreatIntelligenceSystem creates a new threat intelligence system
func NewThreatIntelligenceSystem(config IntelConfig) *ThreatIntelligenceSystem {
    return &ThreatIntelligenceSystem{
        feeds:           make(map[string]*ThreatFeed),
        indicators:      make(map[string]*ThreatIndicator),
        vulnerabilities: make(map[string]*VulnerabilityIntel),
        actors:          make(map[string]*ThreatActor),
        campaigns:       make(map[string]*ThreatCampaign),
        analyzer:        NewIntelligenceAnalyzer(),
        correlator:      NewThreatCorrelator(),
        predictor:       NewThreatPredictor(),
        repository:      NewIntelRepository(),
        config:          config,
    }

// AddFeed adds a threat intelligence feed
}
func (tis *ThreatIntelligenceSystem) AddFeed(ctx context.Context, feed *ThreatFeed) error {
    tis.mu.Lock()
    defer tis.mu.Unlock()

    if len(tis.feeds) >= tis.config.MaxFeeds {
        return fmt.Errorf("maximum feeds limit reached")
    }

    feed.ID = generateFeedID()
    feed.Status = FeedActive
    feed.LastUpdate = time.Now()

    tis.feeds[feed.ID] = feed

    // Start feed monitoring
    go tis.monitorFeed(ctx, feed)

    return nil

// ImportIndicators imports threat indicators
}
func (tis *ThreatIntelligenceSystem) ImportIndicators(ctx context.Context, indicators []*ThreatIndicator) error {
    tis.mu.Lock()
    defer tis.mu.Unlock()

    for _, indicator := range indicators {
        if indicator.ID == "" {
            indicator.ID = generateIndicatorID()
        }
        
        // Validate indicator
        if err := tis.validateIndicator(indicator); err != nil {
            return fmt.Errorf("invalid indicator %s: %w", indicator.ID, err)
        }

        tis.indicators[indicator.ID] = indicator

        // Store in repository
        tis.repository.Store("indicator", indicator.ID, indicator)
    }

    // Trigger correlation if enabled
    if tis.config.AutoCorrelation {
        go tis.correlator.CorrelateIndicators(indicators)
    }

    return nil

// NewIntelligenceAnalyzer creates a new intelligence analyzer
}
func NewIntelligenceAnalyzer() *IntelligenceAnalyzer {
    return &IntelligenceAnalyzer{
        patterns:    make(map[string]*Pattern),
        trends:      make(map[string]*Trend),
        predictions: make(map[string]*Prediction),
    }

// AnalyzeThreatLandscape analyzes the threat landscape
}
func (ia *IntelligenceAnalyzer) AnalyzeThreatLandscape(ctx context.Context, data map[string]interface{}) (*ThreatLandscape, error) {
    landscape := &ThreatLandscape{
        Timestamp:    time.Now(),
        RiskLevel:    calculateOverallRisk(data),
        TopThreats:   identifyTopThreats(data),
        EmergingRisks: identifyEmergingRisks(data),
        Trends:       ia.analyzeTrends(data),
        Predictions:  ia.generatePredictions(data),
    }

    return landscape, nil

// ThreatLandscape represents the current threat landscape
type ThreatLandscape struct {
    Timestamp     time.Time              `json:"timestamp"`
    RiskLevel     RiskLevel              `json:"risk_level"`
    TopThreats    []ThreatSummary        `json:"top_threats"`
    EmergingRisks []EmergingRisk         `json:"emerging_risks"`
    Trends        []Trend                `json:"trends"`
    Predictions   []Prediction           `json:"predictions"`
    Metadata      map[string]interface{} `json:"metadata"`
}

}
// RiskLevel defines risk levels
type RiskLevel string

const (
    RiskCritical RiskLevel = "critical"
    RiskHigh     RiskLevel = "high"
    RiskMedium   RiskLevel = "medium"
    RiskLow      RiskLevel = "low"
)

// ThreatSummary summarizes a threat
type ThreatSummary struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Type        string        `json:"type"`
    Severity    SeverityLevel `json:"severity"`
    Prevalence  float64       `json:"prevalence"`
    Impact      string        `json:"impact"`

}
// EmergingRisk represents an emerging risk
type EmergingRisk struct {
    ID          string          `json:"id"`
    Description string          `json:"description"`
    Indicators  []string        `json:"indicators"`
    Likelihood  float64         `json:"likelihood"`
    FirstSeen   time.Time       `json:"first_seen"`
}

}
// NewThreatCorrelator creates a new threat correlator
func NewThreatCorrelator() *ThreatCorrelator {
    return &ThreatCorrelator{
        correlations: make(map[string]*Correlation),
        rules:        make(map[string]*CorrelationRule),
    }

// CorrelateIndicators correlates threat indicators
}
func (tc *ThreatCorrelator) CorrelateIndicators(indicators []*ThreatIndicator) {
    tc.mu.Lock()
    defer tc.mu.Unlock()

    // Implement correlation logic
    for i, indicator1 := range indicators {
        for j, indicator2 := range indicators {
            if i >= j {
                continue
            }

            strength := tc.calculateCorrelationStrength(indicator1, indicator2)
            if strength > 0.7 {
                correlation := &Correlation{
                    ID:   generateCorrelationID(),
                    Type: CorrelationActorIOC,
                    Entities: []CorrelatedEntity{
                        {Type: "indicator", ID: indicator1.ID, Weight: strength},
                        {Type: "indicator", ID: indicator2.ID, Weight: strength},
                    },
                    Strength:  strength,
                    CreatedAt: time.Now(),
                }
                tc.correlations[correlation.ID] = correlation
            }
        }
    }

// calculateCorrelationStrength calculates correlation strength
}
func (tc *ThreatCorrelator) calculateCorrelationStrength(i1, i2 *ThreatIndicator) float64 {
    // Implement correlation algorithm
    score := 0.0
    
    // Check pattern similarity
    if i1.Pattern != "" && i2.Pattern != "" {
        // Calculate pattern similarity
        score += 0.3
    }
    
    // Check temporal proximity
    timeDiff := i2.FirstSeen.Sub(i1.FirstSeen).Abs()
    if timeDiff < 24*time.Hour {
        score += 0.3
    }
    
    // Check tag overlap
    tagOverlap := calculateTagOverlap(i1.Tags, i2.Tags)
    score += tagOverlap * 0.4
    
    return score

// NewThreatPredictor creates a new threat predictor
}
func NewThreatPredictor() *ThreatPredictor {
    return &ThreatPredictor{
        models:     make(map[string]*PredictionModel),
        historical: &HistoricalData{},
    }

// PredictThreats predicts future threats
}
func (tp *ThreatPredictor) PredictThreats(ctx context.Context, timeframe time.Duration) ([]*Prediction, error) {
    tp.mu.RLock()
    defer tp.mu.RUnlock()

    var predictions []*Prediction

    // Use models to generate predictions
    for _, model := range tp.models {
        prediction := tp.generatePrediction(model, timeframe)
        if prediction != nil {
            predictions = append(predictions, prediction)
        }
    }

    return predictions, nil

// generatePrediction generates a prediction using a model
}
func (tp *ThreatPredictor) generatePrediction(model *PredictionModel, timeframe time.Duration) *Prediction {
    // Implement prediction logic
    return &Prediction{
        ID:          generatePredictionID(),
        Type:        PredictionAttack,
        Description: "Predicted increase in prompt injection attacks",
        Probability: 0.75,
        Impact:      SeverityHigh,
        TimeFrame:   timeframe.String(),
        Confidence:  ConfidenceHigh,
        CreatedAt:   time.Now(),
        ValidUntil:  time.Now().Add(timeframe),
    }

// NewIntelRepository creates a new intelligence repository
}
func NewIntelRepository() *IntelRepository {
    return &IntelRepository{
        storage: make(map[string]interface{}),
        indices: make(map[string]map[string]string),
    }

// Store stores intelligence data
}
func (ir *IntelRepository) Store(dataType, id string, data interface{}) error {
    ir.mu.Lock()
    defer ir.mu.Unlock()

    key := fmt.Sprintf("%s:%s", dataType, id)
    ir.storage[key] = data

    // Update indices
    if _, exists := ir.indices[dataType]; !exists {
        ir.indices[dataType] = make(map[string]string)
    }
    ir.indices[dataType][id] = key

    return nil

// Helper functions
}
func (tis *ThreatIntelligenceSystem) monitorFeed(ctx context.Context, feed *ThreatFeed) {
    ticker := time.NewTicker(feed.Frequency)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Update feed
            if err := tis.updateFeed(feed); err != nil {
                feed.Status = FeedError
            }
        case <-ctx.Done():
            return
        }
    }

}
func (tis *ThreatIntelligenceSystem) updateFeed(feed *ThreatFeed) error {
    // Implement feed update logic
    feed.LastUpdate = time.Now()
    feed.Status = FeedActive
    return nil

func (tis *ThreatIntelligenceSystem) validateIndicator(indicator *ThreatIndicator) error {
    if indicator.Type == "" {
        return fmt.Errorf("indicator type is required")
    }
    if indicator.Value == "" && indicator.Pattern == "" {
        return fmt.Errorf("indicator must have value or pattern")
    }
    return nil

func (ia *IntelligenceAnalyzer) analyzeTrends(data map[string]interface{}) []Trend {
    // Implement trend analysis
    return []Trend{}

}
func (ia *IntelligenceAnalyzer) generatePredictions(data map[string]interface{}) []Prediction {
    // Implement prediction generation
    return []Prediction{}

}
func calculateOverallRisk(data map[string]interface{}) RiskLevel {
    // Implement risk calculation
    return RiskMedium

func identifyTopThreats(data map[string]interface{}) []ThreatSummary {
    // Implement threat identification
    return []ThreatSummary{}

}
func identifyEmergingRisks(data map[string]interface{}) []EmergingRisk {
    // Implement emerging risk identification
    return []EmergingRisk{}

}
func calculateTagOverlap(tags1, tags2 []string) float64 {
    if len(tags1) == 0 || len(tags2) == 0 {
        return 0
    }

    tagMap := make(map[string]bool)
    for _, tag := range tags1 {
        tagMap[tag] = true
    }

    overlap := 0
    for _, tag := range tags2 {
        if tagMap[tag] {
            overlap++
        }
    }

    return float64(overlap) / float64(max(len(tags1), len(tags2)))

}
func max(a, b int) int {
    if a > b {
        return a
    }
    return b

func generateFeedID() string {
    return fmt.Sprintf("feed_%d", time.Now().UnixNano())

}
func generateIndicatorID() string {
    return fmt.Sprintf("indicator_%d", time.Now().UnixNano())

}
func generateCorrelationID() string {
    return fmt.Sprintf("corr_%d", time.Now().UnixNano())

}
func generatePredictionID() string {
    return fmt.Sprintf("pred_%d", time.Now().UnixNano())

// SearchIndicators searches for threat indicators
}
func (tis *ThreatIntelligenceSystem) SearchIndicators(ctx context.Context, query string) ([]*ThreatIndicator, error) {
    tis.mu.RLock()
    defer tis.mu.RUnlock()

    var results []*ThreatIndicator
    
    for _, indicator := range tis.indicators {
        if matchesQuery(indicator, query) {
            results = append(results, indicator)
        }
    }

    return results, nil

func matchesQuery(indicator *ThreatIndicator, query string) bool {
    // Implement query matching logic
    return contains(indicator.Description, query) || 
           contains(indicator.Value, query) ||
           contains(indicator.Pattern, query)

}
func contains(s, substr string) bool {
    return len(s) >= len(substr) && s == substr || 
           len(s) > len(substr) && (s[:len(substr)] == substr || 
           s[len(s)-len(substr):] == substr || 
           findSubstring(s, substr) != -1)

}
func findSubstring(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
