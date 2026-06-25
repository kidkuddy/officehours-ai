// Package models holds Go structs mirroring the Postgres schema in
// backend/migrations/0001_init.sql. Frozen field/table names per BUILD_SPEC §2.
package models

import "time"

// Stages are the 6 maturity stages (exact strings, frozen).
const (
	StageIdeation         = "Ideation"
	StageMarketValidation = "Market Validation"
	StageStructuration    = "Structuration"
	StageFundraising      = "Fundraising"
	StageLaunchPlanning   = "Launch Planning"
	StageGrowth           = "Growth"
)

// Stages is the ordered list of the 6 maturity stages.
var Stages = []string{
	StageIdeation,
	StageMarketValidation,
	StageStructuration,
	StageFundraising,
	StageLaunchPlanning,
	StageGrowth,
}

// Signal names (exact strings, frozen).
const (
	SignalMarket          = "Market"
	SignalCommercialOffer = "Commercial Offer"
	SignalInnovation      = "Innovation"
	SignalScalability     = "Scalability"
	SignalGreen           = "Green"
)

// SignalNames is the ordered list of the 5 composite Signals.
var SignalNames = []string{
	SignalMarket,
	SignalCommercialOffer,
	SignalInnovation,
	SignalScalability,
	SignalGreen,
}

// Session kinds.
const (
	SessionKindOfficeHours = "office_hours"
	SessionKindLearn       = "learn"
)

// Session statuses.
const (
	SessionStatusActive    = "active"
	SessionStatusConcluded = "concluded"
)

// Agent job types.
const (
	JobTypeAdvisor   = "advisor"
	JobTypeDiagnoser = "diagnoser"
	JobTypeScorer    = "scorer"
	JobTypeLearn     = "learn"
)

// Agent job statuses.
const (
	JobStatusQueued  = "queued"
	JobStatusRunning = "running"
	JobStatusDone    = "done"
	JobStatusError   = "error"
)

// User mirrors the users table.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

// Profile mirrors the profiles table.
type Profile struct {
	UserID        string     `json:"user_id"`
	CompanyText   string     `json:"company_text"`
	Stage         string     `json:"stage"`
	StageEvidence []byte     `json:"stage_evidence"` // raw JSON array
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	OnboardedAt   *time.Time `json:"onboarded_at,omitempty"`
}

// Subscore is one weighted criterion contributing to a Signal.
type Subscore struct {
	Criterion    string  `json:"criterion"`
	Weight       float64 `json:"weight"`
	Score        float64 `json:"score"`
	Contribution float64 `json:"contribution"`
}

// Signal mirrors the signals table.
type Signal struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	Score          float64   `json:"score"`
	Subscores      []byte    `json:"subscores"` // raw JSON array of Subscore
	Rationale      string    `json:"rationale"`
	FloorTriggered bool      `json:"floor_triggered"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Session mirrors the sessions table.
type Session struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Kind        string     `json:"kind"`
	AdvisorKey  string     `json:"advisor_key"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Outcomes    string     `json:"outcomes"`
	CreatedAt   time.Time  `json:"created_at"`
	ConcludedAt *time.Time `json:"concluded_at,omitempty"`
}

// Message mirrors the messages table.
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Goal mirrors the goals table.
type Goal struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	SourceSessionID *string    `json:"source_session_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	DoneAt          *time.Time `json:"done_at,omitempty"`
}

// ActionItem mirrors the action_items table.
type ActionItem struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	SessionID  *string   `json:"session_id,omitempty"`
	Title      string    `json:"title"`
	Horizon    string    `json:"horizon"`
	Rationale  string    `json:"rationale"`
	ProgramRef string    `json:"program_ref"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// Document mirrors the documents table.
type Document struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"` // null => shared KB
	Collection string    `json:"collection"`
	Filename   string    `json:"filename"`
	Mime       string    `json:"mime"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"created_at"`
}

// Chunk mirrors the chunks table (tsv column not surfaced to Go).
type Chunk struct {
	ID         string  `json:"id"`
	DocumentID string  `json:"document_id"`
	Collection string  `json:"collection"`
	UserID     *string `json:"user_id,omitempty"`
	Ord        int     `json:"ord"`
	Content    string  `json:"content"`
}

// Event mirrors the events table (Mon Parcours timeline).
type Event struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	Payload   []byte    `json:"payload"` // raw JSON object
	CreatedAt time.Time `json:"created_at"`
}

// AgentJob mirrors the agent_jobs table.
type AgentJob struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	UserID     string     `json:"user_id"`
	SessionID  *string    `json:"session_id,omitempty"`
	Status     string     `json:"status"`
	Input      []byte     `json:"input"`  // raw JSON object
	Output     []byte     `json:"output"` // raw JSON object
	Error      string     `json:"error"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
