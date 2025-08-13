package collaboration

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
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

// CollaborationSession represents an active collaboration session
type CollaborationSession struct {
    ID           string                `json:"id"`
    ProjectID    string                `json:"project_id"`
    Participants []*SessionParticipant `json:"participants"`
    SharedState  map[string]interface{} `json:"shared_state"`
    Activities   []*Activity           `json:"activities"`
    StartedAt    time.Time            `json:"started_at"`
    LastActivity time.Time            `json:"last_activity"`
}

// SessionParticipant represents a session participant
type SessionParticipant struct {
    UserID     string    `json:"user_id"`
    Username   string    `json:"username"`
    Role       string    `json:"role"`
    JoinedAt   time.Time `json:"joined_at"`
    Active     bool      `json:"active"`
    Cursor     *Cursor   `json:"cursor,omitempty"`
}

// Cursor represents user cursor position in shared workspace
type Cursor struct {
    X        int    `json:"x"`
    Y        int    `json:"y"`
    View     string `json:"view"`
    Selected string `json:"selected"`
}

// Activity represents a user activity
type Activity struct {
    ID          string                 `json:"id"`
    UserID      string                 `json:"user_id"`
    Type        ActivityType           `json:"type"`
    Description string                 `json:"description"`
    Target      string                 `json:"target"`
    Metadata    map[string]interface{} `json:"metadata"`
    Timestamp   time.Time             `json:"timestamp"`
}

// ActivityType defines activity types
type ActivityType string

const (
    ActivityCreate   ActivityType = "create"
    ActivityUpdate   ActivityType = "update"
    ActivityDelete   ActivityType = "delete"
    ActivityExecute  ActivityType = "execute"
    ActivityComment  ActivityType = "comment"
    ActivityShare    ActivityType = "share"
)

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

// CreateProject creates a new project
func (cp *CollaborationPlatform) CreateProject(ctx context.Context, teamID string, projectDef *Project) (*Project, error) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    team, exists := cp.teams[teamID]
    if !exists {
        return nil, fmt.Errorf("team not found")
    }

    if len(team.Projects) >= cp.config.MaxProjectsPerTeam {
        return nil, fmt.Errorf("maximum projects limit reached for team")
    }

    project := &Project{
        ID:              generateProjectID(),
        Name:            projectDef.Name,
        Description:     projectDef.Description,
        TeamID:          teamID,
        Target:          projectDef.Target,
        Scope:           projectDef.Scope,
        Timeline:        projectDef.Timeline,
        Findings:        []*Finding{},
        SharedResources: make(map[string]*Resource),
        Status:          ProjectPlanning,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    cp.projects[project.ID] = project
    team.Projects = append(team.Projects, project.ID)

    // Initialize project workspace
    cp.workspace.CreateProjectSpace(project.ID)

    // Create initial project structure
    cp.initializeProjectStructure(project)

    return project, nil
}

// StartCollaborationSession starts a new collaboration session
func (cp *CollaborationPlatform) StartCollaborationSession(ctx context.Context, projectID string, initiator *TeamMember) (*CollaborationSession, error) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    project, exists := cp.projects[projectID]
    if !exists {
        return nil, fmt.Errorf("project not found")
    }

    session := &CollaborationSession{
        ID:        generateSessionID(),
        ProjectID: projectID,
        Participants: []*SessionParticipant{{
            UserID:   initiator.ID,
            Username: initiator.Username,
            Role:     string(initiator.Role),
            JoinedAt: time.Now(),
            Active:   true,
        }},
        SharedState:  make(map[string]interface{}),
        Activities:   []*Activity{},
        StartedAt:    time.Now(),
        LastActivity: time.Now(),
    }

    cp.sessions[session.ID] = session

    // Notify team members
    cp.messenger.BroadcastToTeam(project.TeamID, &Message{
        Type:    "session_started",
        Content: fmt.Sprintf("%s started a collaboration session for %s", initiator.Username, project.Name),
        Data:    map[string]string{"session_id": session.ID},
    })

    return session, nil
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

// SharedFile represents a shared file
type SharedFile struct {
    ID       string
    Name     string
    Content  []byte
    Version  int
    LockedBy string
    History  []*FileVersion
}

// SharedBoard represents a shared planning board
type SharedBoard struct {
    ID       string
    Name     string
    Elements map[string]*BoardElement
}

// AttackPlan represents an attack plan
type AttackPlan struct {
    ID          string
    Name        string
    Description string
    Steps       []*AttackStep
    Status      string
}

// TestResult represents test results
type TestResult struct {
    ID        string
    TestName  string
    Success   bool
    Output    string
    Artifacts []string
}

// Note represents a shared note
type Note struct {
    ID        string
    Title     string
    Content   string
    Author    string
    CreatedAt time.Time
    Tags      []string
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

// initializeProjectStructure sets up initial project structure
func (cp *CollaborationPlatform) initializeProjectStructure(project *Project) {
    // Create default resources
    defaultResources := []Resource{
        {
            ID:   generateResourceID(),
            Name: "Attack Playbook",
            Type: ResourceTemplate,
            Content: "# Attack Playbook\n\n## Objectives\n\n## Techniques\n\n## Timeline",
        },
        {
            ID:   generateResourceID(),
            Name: "Payload Library",
            Type: ResourcePayload,
            Content: "// Payload collection for " + project.Name,
        },
        {
            ID:   generateResourceID(),
            Name: "Test Scripts",
            Type: ResourceScript,
            Content: "# Test automation scripts",
        },
    }

    for _, resource := range defaultResources {
        resource.Owner = "system"
        resource.Version = 1
        resource.UpdatedAt = time.Now()
        project.SharedResources[resource.ID] = &resource
    }
}

func generateResourceID() string {
    return fmt.Sprintf("res_%d", time.Now().UnixNano())
}

// JoinSession allows a user to join a collaboration session
func (cp *CollaborationPlatform) JoinSession(ctx context.Context, sessionID string, member *TeamMember) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    session, exists := cp.sessions[sessionID]
    if !exists {
        return fmt.Errorf("session not found")
    }

    // Check if already in session
    for _, p := range session.Participants {
        if p.UserID == member.ID {
            p.Active = true
            return nil
        }
    }

    // Add new participant
    participant := &SessionParticipant{
        UserID:   member.ID,
        Username: member.Username,
        Role:     string(member.Role),
        JoinedAt: time.Now(),
        Active:   true,
    }

    session.Participants = append(session.Participants, participant)
    session.LastActivity = time.Now()

    // Record activity
    activity := &Activity{
        ID:          generateActivityID(),
        UserID:      member.ID,
        Type:        "join_session",
        Description: fmt.Sprintf("%s joined the session", member.Username),
        Timestamp:   time.Now(),
    }
    session.Activities = append(session.Activities, activity)

    return nil
}

func generateActivityID() string {
    return fmt.Sprintf("act_%d", time.Now().UnixNano())
}

// AddFinding adds a finding to a project
func (cp *CollaborationPlatform) AddFinding(ctx context.Context, projectID string, finding *Finding) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    project, exists := cp.projects[projectID]
    if !exists {
        return fmt.Errorf("project not found")
    }

    finding.ID = generateFindingID()
    finding.DiscoveredAt = time.Now()
    finding.Status = FindingNew

    project.Findings = append(project.Findings, finding)
    project.UpdatedAt = time.Now()

    // Share with team
    cp.knowledge.ShareFinding(project.TeamID, finding)

    return nil
}

func generateFindingID() string {
    return fmt.Sprintf("finding_%d", time.Now().UnixNano())
}

// ShareResource shares a resource within a project
func (cp *CollaborationPlatform) ShareResource(ctx context.Context, projectID string, resource *Resource) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    project, exists := cp.projects[projectID]
    if !exists {
        return fmt.Errorf("project not found")
    }

    resource.ID = generateResourceID()
    resource.Version = 1
    resource.UpdatedAt = time.Now()

    project.SharedResources[resource.ID] = resource
    project.UpdatedAt = time.Now()

    return nil
}

// ShareFinding shares a finding with the team knowledge base
func (kb *KnowledgeBase) ShareFinding(teamID string, finding *Finding) {
    kb.mu.Lock()
    defer kb.mu.Unlock()

    if teamKB, exists := kb.teamKBs[teamID]; exists {
        // Convert finding to technique if applicable
        technique := &Technique{
            ID:          generateTechniqueID(),
            Name:        finding.Title,
            Description: finding.Description,
            Category:    finding.Category,
            Examples:    []string{finding.Description},
        }
        teamKB.Techniques[technique.ID] = technique
    }
}

func generateTechniqueID() string {
    return fmt.Sprintf("tech_%d", time.Now().UnixNano())
}