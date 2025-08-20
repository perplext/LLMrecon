package collaboration

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// CollaborationPlatform manages team collaboration for LLM red teaming
type CollaborationPlatform struct {
    mu          sync.RWMutex
    teams       map[string]*Team
    projects    map[string]*Project
    sessions    map[string]*CollaborationSession
    workspace   *SharedWorkspace
    messenger   *TeamMessenger
    taskManager *TaskManager
    knowledge   *KnowledgeBase
    config      PlatformConfig
}

// PlatformConfig holds configuration for collaboration platform
type PlatformConfig struct {
    MaxTeams           int
    MaxProjectsPerTeam int
    SessionTimeout     time.Duration
    AutoSyncInterval   time.Duration
    EncryptionEnabled  bool
}

// Team represents a red team group
type Team struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Members     []*TeamMember       `json:"members"`
    Projects    []string            `json:"projects"`
    Permissions map[string][]string `json:"permissions"`
    CreatedAt   time.Time          `json:"created_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// TeamMember represents a team member
type TeamMember struct {
    ID           string    `json:"id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    Role         TeamRole  `json:"role"`
    Specialties  []string  `json:"specialties"`
    Status       UserStatus `json:"status"`
    LastActive   time.Time `json:"last_active"`
    Contributions int      `json:"contributions"`
}

// TeamRole defines team member roles
type TeamRole string

const (
    RoleAdmin      TeamRole = "admin"
    RoleLeader     TeamRole = "leader"
    RoleResearcher TeamRole = "researcher"
    RoleAnalyst    TeamRole = "analyst"
    RoleEngineer   TeamRole = "engineer"
    RoleObserver   TeamRole = "observer"
)

// UserStatus defines user status
type UserStatus string

const (
    StatusOnline    UserStatus = "online"
    StatusAway      UserStatus = "away"
    StatusBusy      UserStatus = "busy"
    StatusOffline   UserStatus = "offline"
)

// Project represents a red team project
type Project struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    TeamID          string                 `json:"team_id"`
    Target          *TargetSystem          `json:"target"`
    Scope           *ProjectScope          `json:"scope"`
    Timeline        *ProjectTimeline       `json:"timeline"`
    Findings        []*Finding             `json:"findings"`
    SharedResources map[string]*Resource   `json:"shared_resources"`
    Status          ProjectStatus          `json:"status"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}

// TargetSystem represents the LLM system being tested
type TargetSystem struct {
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`
    Version      string                 `json:"version"`
    Endpoints    []string               `json:"endpoints"`
    Credentials  map[string]string      `json:"credentials"`
    Constraints  []string               `json:"constraints"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// ProjectScope defines the scope of testing
type ProjectScope struct {
    InScope      []string `json:"in_scope"`
    OutOfScope   []string `json:"out_of_scope"`
    Objectives   []string `json:"objectives"`
    Limitations  []string `json:"limitations"`
    RulesOfEngagement string `json:"rules_of_engagement"`
}

// ProjectTimeline defines project timeline
type ProjectTimeline struct {
    StartDate    time.Time `json:"start_date"`
    EndDate      time.Time `json:"end_date"`
    Milestones   []*Milestone `json:"milestones"`
    CurrentPhase string    `json:"current_phase"`
}

// Milestone represents a project milestone
type Milestone struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    DueDate     time.Time `json:"due_date"`
    Completed   bool      `json:"completed"`
    CompletedAt time.Time `json:"completed_at"`
}

// ProjectStatus defines project status
type ProjectStatus string

const (
    ProjectPlanning    ProjectStatus = "planning"
    ProjectActive      ProjectStatus = "active"
    ProjectPaused      ProjectStatus = "paused"
    ProjectCompleted   ProjectStatus = "completed"
    ProjectArchived    ProjectStatus = "archived"
)

// Finding represents a discovered vulnerability or issue
type Finding struct {
    ID           string         `json:"id"`
    Title        string         `json:"title"`
    Description  string         `json:"description"`
    Severity     SeverityLevel  `json:"severity"`
    Category     string         `json:"category"`
    Evidence     []*Evidence    `json:"evidence"`
    Reproducible bool           `json:"reproducible"`
    DiscoveredBy string         `json:"discovered_by"`
    DiscoveredAt time.Time      `json:"discovered_at"`
    Status       FindingStatus  `json:"status"`
}

// SeverityLevel defines finding severity
type SeverityLevel string

const (
    SeverityCritical SeverityLevel = "critical"
    SeverityHigh     SeverityLevel = "high"
    SeverityMedium   SeverityLevel = "medium"
    SeverityLow      SeverityLevel = "low"
    SeverityInfo     SeverityLevel = "info"
)

// FindingStatus defines finding status
type FindingStatus string

const (
    FindingNew        FindingStatus = "new"
    FindingConfirmed  FindingStatus = "confirmed"
    FindingInProgress FindingStatus = "in_progress"
    FindingResolved   FindingStatus = "resolved"
    FindingFalsePositive FindingStatus = "false_positive"
)

// Evidence represents evidence for a finding
type Evidence struct {
    Type        string                 `json:"type"`
    Data        string                 `json:"data"`
    Screenshot  []byte                 `json:"screenshot,omitempty"`
    Timestamp   time.Time             `json:"timestamp"`
    CollectedBy string                 `json:"collected_by"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Resource represents a shared resource
type Resource struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Type        ResourceType `json:"type"`
    Content     string       `json:"content"`
    Owner       string       `json:"owner"`
    Permissions []string     `json:"permissions"`
    Version     int          `json:"version"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

// ResourceType defines resource types
type ResourceType string

const (
    ResourcePayload    ResourceType = "payload"
    ResourceScript     ResourceType = "script"
    ResourceTemplate   ResourceType = "template"
    ResourceReport     ResourceType = "report"
    ResourceNote       ResourceType = "note"
    ResourceTool       ResourceType = "tool"
)

// CollaborationSession represents a collaboration session
type CollaborationSession struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    TeamID      string                 `json:"team_id"`
    ProjectID   string                 `json:"project_id"`
    Participants []string              `json:"participants"`
    StartTime   time.Time              `json:"start_time"`
    EndTime     *time.Time             `json:"end_time,omitempty"`
    Status      SessionStatus          `json:"status"`
    Activities  []*SessionActivity     `json:"activities"`
    SharedFiles []string               `json:"shared_files"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// SessionStatus defines session status
type SessionStatus string

const (
    SessionActive    SessionStatus = "active"
    SessionPaused    SessionStatus = "paused"
    SessionCompleted SessionStatus = "completed"
    SessionAbandoned SessionStatus = "abandoned"
)

// SessionActivity represents an activity within a session
type SessionActivity struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    PerformedBy string                 `json:"performed_by"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        map[string]interface{} `json:"data"`
}

// NewCollaborationPlatform creates a new collaboration platform
func NewCollaborationPlatform(config PlatformConfig) *CollaborationPlatform {
    return &CollaborationPlatform{
        teams:       make(map[string]*Team),
        projects:    make(map[string]*Project),
        sessions:    make(map[string]*CollaborationSession),
        workspace:   NewSharedWorkspace(),
        messenger:   NewTeamMessenger(),
        taskManager: NewTaskManager(),
        knowledge:   NewKnowledgeBase(),
        config:      config,
    }
}

// CreateTeam creates a new team
func (cp *CollaborationPlatform) CreateTeam(ctx context.Context, name string, admin *TeamMember) (*Team, error) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if len(cp.teams) >= cp.config.MaxTeams {
        return nil, fmt.Errorf("maximum teams limit reached")
    }

    team := &Team{
        ID:          generateTeamID(),
        Name:        name,
        Members:     []*TeamMember{admin},
        Projects:    []string{},
        Permissions: make(map[string][]string),
        CreatedAt:   time.Now(),
        Metadata:    make(map[string]interface{}),
    }

    // Set admin permissions
    team.Permissions[admin.ID] = []string{"all"}
    admin.Role = RoleAdmin

    cp.teams[team.ID] = team
    
    // Create team workspace
    cp.workspace.CreateTeamSpace(team.ID)
    
    // Initialize team knowledge base
    cp.knowledge.CreateTeamKB(team.ID)

    return team, nil
}

// SharedWorkspace manages shared workspace functionality
type SharedWorkspace struct {
    teamSpaces    map[string]*TeamSpace
    projectSpaces map[string]*ProjectSpace
    mu            sync.RWMutex
}

// TeamSpace represents a team's shared workspace
type TeamSpace struct {
    TeamID       string
    SharedFiles  map[string]*SharedFile
    SharedBoards map[string]*SharedBoard
}

// ProjectSpace represents a project's workspace
type ProjectSpace struct {
    ProjectID    string
    AttackPlans  map[string]*AttackPlan
    TestResults  map[string]*TestResult
    SharedNotes  map[string]*Note
}

// NewSharedWorkspace creates a new shared workspace
func NewSharedWorkspace() *SharedWorkspace {
    return &SharedWorkspace{
        teamSpaces:    make(map[string]*TeamSpace),
        projectSpaces: make(map[string]*ProjectSpace),
    }
}

// CreateTeamSpace creates a team workspace
func (sw *SharedWorkspace) CreateTeamSpace(teamID string) {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    sw.teamSpaces[teamID] = &TeamSpace{
        TeamID:       teamID,
        SharedFiles:  make(map[string]*SharedFile),
        SharedBoards: make(map[string]*SharedBoard),
    }
}

// CreateProjectSpace creates a project workspace
func (sw *SharedWorkspace) CreateProjectSpace(projectID string) {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    sw.projectSpaces[projectID] = &ProjectSpace{
        ProjectID:   projectID,
        AttackPlans: make(map[string]*AttackPlan),
        TestResults: make(map[string]*TestResult),
        SharedNotes: make(map[string]*Note),
    }
}

// TeamMessenger handles team communication
type TeamMessenger struct {
    channels map[string]*MessageChannel
    mu       sync.RWMutex
}

// MessageChannel represents a message channel
type MessageChannel struct {
    ID       string
    Type     string
    Members  []string
    Messages []*Message
}

// Message represents a team message
type Message struct {
    ID        string
    Type      string
    Content   string
    From      string
    To        []string
    Data      map[string]string
    Timestamp time.Time
}

// NewTeamMessenger creates a new team messenger
func NewTeamMessenger() *TeamMessenger {
    return &TeamMessenger{
        channels: make(map[string]*MessageChannel),
    }
}

// BroadcastToTeam broadcasts a message to team
func (tm *TeamMessenger) BroadcastToTeam(teamID string, message *Message) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    channelID := fmt.Sprintf("team_%s", teamID)
    channel, exists := tm.channels[channelID]
    if !exists {
        channel = &MessageChannel{
            ID:       channelID,
            Type:     "team",
            Members:  []string{},
            Messages: []*Message{},
        }
        tm.channels[channelID] = channel
    }

    message.ID = generateMessageID()
    message.Timestamp = time.Now()
    channel.Messages = append(channel.Messages, message)

    return nil
}

// TaskManager manages collaborative tasks
type TaskManager struct {
    tasks      map[string]*Task
    assignments map[string][]string
    mu         sync.RWMutex
}

// Task represents a collaborative task
type Task struct {
    ID          string
    Title       string
    Description string
    AssignedTo  []string
    Status      TaskStatus
    Priority    TaskPriority
    DueDate     time.Time
    Dependencies []string
    Subtasks    []*Task
}

// TaskStatus defines task status
type TaskStatus string

const (
    TaskTodo       TaskStatus = "todo"
    TaskInProgress TaskStatus = "in_progress"
    TaskReview     TaskStatus = "review"
    TaskDone       TaskStatus = "done"
)

// TaskPriority defines task priority
type TaskPriority string

const (
    PriorityCritical TaskPriority = "critical"
    PriorityHigh     TaskPriority = "high"
    PriorityMedium   TaskPriority = "medium"
    PriorityLow      TaskPriority = "low"
)

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
    return &TaskManager{
        tasks:       make(map[string]*Task),
        assignments: make(map[string][]string),
    }
}

// KnowledgeBase manages shared knowledge
type KnowledgeBase struct {
    teamKBs    map[string]*TeamKnowledgeBase
    globalKB   *GlobalKnowledgeBase
    mu         sync.RWMutex
}

// TeamKnowledgeBase represents team-specific knowledge
type TeamKnowledgeBase struct {
    TeamID     string
    Techniques map[string]*Technique
    Payloads   map[string]*SavedPayload
    Reports    map[string]*Report
}

// GlobalKnowledgeBase represents global knowledge
type GlobalKnowledgeBase struct {
    CommonVulnerabilities map[string]*Vulnerability
    BestPractices        map[string]*Practice
    Tools                map[string]*Tool
}

// Technique represents an attack technique
type Technique struct {
    ID          string
    Name        string
    Description string
    Category    string
    Steps       []string
    Examples    []string
    References  []string
}

// SavedPayload represents a saved payload
type SavedPayload struct {
    ID          string
    Name        string
    Description string
    Payload     string
    SuccessRate float64
    Tags        []string
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase() *KnowledgeBase {
    return &KnowledgeBase{
        teamKBs: make(map[string]*TeamKnowledgeBase),
        globalKB: &GlobalKnowledgeBase{
            CommonVulnerabilities: make(map[string]*Vulnerability),
            BestPractices:        make(map[string]*Practice),
            Tools:                make(map[string]*Tool),
        },
    }
}

// CreateTeamKB creates a team knowledge base
func (kb *KnowledgeBase) CreateTeamKB(teamID string) {
    kb.mu.Lock()
    defer kb.mu.Unlock()

    kb.teamKBs[teamID] = &TeamKnowledgeBase{
        TeamID:     teamID,
        Techniques: make(map[string]*Technique),
        Payloads:   make(map[string]*SavedPayload),
        Reports:    make(map[string]*Report),
    }
}

// Helper types
type Vulnerability struct {
    ID          string
    Name        string
    Description string
    Severity    string
    Mitigation  string
}

type Practice struct {
    ID          string
    Title       string
    Description string
    Category    string
}

type Tool struct {
    ID          string
    Name        string
    Description string
    Usage       string
}

type Report struct {
    ID       string
    Title    string
    Content  string
    Author   string
    Date     time.Time
}

type SharedFile struct {
    ID       string
    Name     string
    Content  []byte
    Version  int
    LockedBy string
    History  []*FileVersion
}

type SharedBoard struct {
    ID       string
    Name     string
    Elements map[string]*BoardElement
}

type AttackPlan struct {
    ID          string
    Name        string
    Description string
    Steps       []*AttackStep
    Status      string
}

type TestResult struct {
    ID        string
    TestName  string
    Success   bool
    Output    string
    Artifacts []string
}

type Note struct {
    ID        string
    Title     string
    Content   string
    Author    string
    CreatedAt time.Time
    Tags      []string
}

type FileVersion struct {
    Version   int
    Content   []byte
    Author    string
    Timestamp time.Time
}

type BoardElement struct {
    ID       string
    Type     string
    Position Position
    Content  string
}

type Position struct {
    X int
    Y int
}

type AttackStep struct {
    Order       int
    Description string
    Completed   bool
}

// Helper functions
func generateTeamID() string {
    return fmt.Sprintf("team_%d", time.Now().UnixNano())
}

func generateProjectID() string {
    return fmt.Sprintf("proj_%d", time.Now().UnixNano())
}

func generateSessionID() string {
    return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

func generateMessageID() string {
    return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func generateResourceID() string {
    return fmt.Sprintf("res_%d", time.Now().UnixNano())
}

func generateActivityID() string {
    return fmt.Sprintf("act_%d", time.Now().UnixNano())
}

func generateFindingID() string {
    return fmt.Sprintf("finding_%d", time.Now().UnixNano())
}

func generateTechniqueID() string {
    return fmt.Sprintf("tech_%d", time.Now().UnixNano())
}