package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	CaseSessionStatusActive    = "active"
	CaseSessionStatusCompleted = "completed"
	CaseSessionStatusAbandoned = "abandoned"

	CaseSessionEventSessionStarted = "session_started"
	CaseSessionEventEvidenceOpened = "evidence_opened"
	CaseSessionEventAnswerSaved    = "answer_saved"
	CaseSessionEventSubmitted      = "session_submitted"

	InitialAssessmentTrusted    = "trusted"
	InitialAssessmentNeedCheck  = "need_check"
	InitialAssessmentMisleading = "misleading"
)

type GetCaseSessionParam struct {
	CaseSessionID         uuid.UUID
	UserID                uuid.UUID
	CaseID                uuid.UUID
	CaseVersionID         uuid.UUID
	Status                string
	ActiveSessionKey      string
	StartIdempotencyKey   string
	IncludeDeletedRecords bool
}

type ListCaseSessionProgressParam struct {
	CaseSessionID uuid.UUID
	UserID        uuid.UUID
}

type CountCaseSessionsStartedParam struct {
	UserID    uuid.UUID
	StartAt   time.Time
	EndAt     time.Time
	Statuses  []string
	ExcludeID uuid.UUID
}

type ListRecentCaseSessionsParam struct {
	UserID uuid.UUID
	CaseID uuid.UUID
	Status string
	Limit  int
}

type GetCaseSessionAnswerParam struct {
	CaseSessionID  uuid.UUID
	CaseQuestionID uuid.UUID
}

type ListCaseSessionAnswersParam struct {
	CaseSessionID uuid.UUID
	IsFinal       *bool
}

type GetCaseSessionEvidenceProgressParam struct {
	CaseSessionID  uuid.UUID
	CaseEvidenceID uuid.UUID
}

type ListCaseSessionEvidenceProgressParam struct {
	CaseSessionID uuid.UUID
}

type ListCaseSessionEventsParam struct {
	CaseSessionID  uuid.UUID
	EventType      string
	CaseEvidenceID uuid.UUID
	CaseQuestionID uuid.UUID
	Limit          int
	Offset         int
}

type GetCaseSessionIdempotencyKeyParam struct {
	IdempotencyKey string
	UserID         uuid.UUID
	RequestMethod  string
	RequestPath    string
	Now            time.Time
}

type StartCaseSessionRequest struct {
	InitialAssessment *string `json:"initial_assessment"`
	InitialConfidence *int    `json:"initial_confidence"`
	AppVersion        *string `json:"app_version"`
}

type SaveCaseSessionAnswersRequest struct {
	SessionVersion int                         `json:"session_version" binding:"required"`
	Answers        []SaveCaseSessionAnswerItem `json:"answers" binding:"required"`
}

type SaveCaseSessionAnswerItem struct {
	CaseQuestionID    uuid.UUID       `json:"case_question_id" binding:"required"`
	QuestionType      string          `json:"question_type" binding:"required"`
	Value             json.RawMessage `json:"value" binding:"required"`
	ConfidenceInitial *int            `json:"confidence_initial"`
	ConfidenceFinal   *int            `json:"confidence_final"`
	IsFinal           bool            `json:"is_final"`
}

type OpenCaseSessionEvidenceRequest struct {
	SessionVersion int `json:"session_version" binding:"required"`
}

type SubmitCaseSessionRequest struct {
	SessionVersion  int    `json:"session_version" binding:"required"`
	FinalDecision   string `json:"final_decision" binding:"required"`
	FinalConfidence int    `json:"final_confidence"`
	Reason          string `json:"reason" binding:"required"`
}

type StartCaseSessionResponse struct {
	Session  GameplaySessionResponse `json:"session"`
	Gameplay GameplayStateResponse   `json:"gameplay"`
}

type GameplayStateResponse struct {
	Session          GameplaySessionResponse            `json:"session"`
	Case             GameplayCaseSnapshotResponse       `json:"case"`
	ChatbotConfig    *GameplayChatbotConfigResponse     `json:"chatbot_config,omitempty"`
	Evidences        []GameplayEvidenceResponse         `json:"evidences"`
	Questions        []GameplayQuestionResponse         `json:"questions"`
	Answers          []GameplayAnswerResponse           `json:"answers"`
	EvidenceProgress []GameplayEvidenceProgressResponse `json:"evidence_progress"`
	Progress         GameplayProgressResponse           `json:"progress"`
}

type GameplaySessionResponse struct {
	CaseSessionID     uuid.UUID  `json:"case_session_id"`
	UserID            uuid.UUID  `json:"user_id"`
	CaseID            uuid.UUID  `json:"case_id"`
	CaseVersionID     uuid.UUID  `json:"case_version_id"`
	SessionVersion    int        `json:"session_version"`
	Status            string     `json:"status"`
	InitialAssessment *string    `json:"initial_assessment"`
	InitialConfidence *int       `json:"initial_confidence"`
	StartedAt         time.Time  `json:"started_at"`
	LastActivityAt    time.Time  `json:"last_activity_at"`
	SubmittedAt       *time.Time `json:"submitted_at"`
}

type GameplayProgressResponse struct {
	OpenedEvidenceCount   int  `json:"opened_evidence_count"`
	TotalEvidenceCount    int  `json:"total_evidence_count"`
	AnsweredQuestionCount int  `json:"answered_question_count"`
	RequiredQuestionCount int  `json:"required_question_count"`
	CanTakeDecision       bool `json:"can_take_decision"`
}

type GameplayEvidenceProgressResponse struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
	Opened         bool      `json:"opened"`
	OpenedCount    int       `json:"opened_count"`
	FirstOpenedAt  time.Time `json:"first_opened_at,omitempty"`
	LastOpenedAt   time.Time `json:"last_opened_at,omitempty"`
}

type GameplayAnswerResponse struct {
	CaseQuestionID    uuid.UUID       `json:"case_question_id"`
	QuestionType      string          `json:"question_type"`
	Value             json.RawMessage `json:"value"`
	ConfidenceInitial *int            `json:"confidence_initial"`
	ConfidenceFinal   *int            `json:"confidence_final"`
	IsFinal           bool            `json:"is_final"`
	SavedAt           time.Time       `json:"saved_at"`
}

type OpenCaseSessionEvidenceResponse struct {
	Session          GameplaySessionResponse          `json:"session"`
	Evidence         GameplayEvidenceResponse         `json:"evidence"`
	EvidenceProgress GameplayEvidenceProgressResponse `json:"evidence_progress"`
	Progress         GameplayProgressResponse         `json:"progress"`
}

type SaveCaseSessionAnswersResponse struct {
	Session  GameplaySessionResponse  `json:"session"`
	Answers  []GameplayAnswerResponse `json:"answers"`
	Progress GameplayProgressResponse `json:"progress"`
}

type SubmitCaseSessionResponse struct {
	Session        GameplaySessionResponse          `json:"session"`
	Outcome        GameplayOutcomeResponse          `json:"outcome"`
	ScoreBreakdown []GameplayScoreBreakdownResponse `json:"score_breakdown"`
	CityImpact     []GameplayCityImpactResponse     `json:"city_impact"`
	Rewards        GameplayRewardResponse           `json:"rewards"`
	Progression    GameplayProgressionResponse      `json:"progression"`
	Feedback       GameplayFeedbackResponse         `json:"feedback"`
}

type GameplayOutcomeResponse struct {
	OutcomeKey   string `json:"outcome_key"`
	OutcomeID    string `json:"outcome_id"`
	OutcomeLabel string `json:"outcome_label"`
	Narrative    string `json:"narrative"`
	TotalScore   int    `json:"total_score"`
}

type GameplayScoreBreakdownResponse struct {
	CategoryKey       string `json:"category_key"`
	CategoryLabel     string `json:"category_label"`
	Score             int    `json:"score"`
	WeightBasisPoints int    `json:"weight_basis_points"`
	WeightedScore     int    `json:"weighted_score"`
}

type GameplayCityImpactResponse struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Delta  int    `json:"delta"`
	Before int    `json:"before"`
	After  int    `json:"after"`
}

type GameplayRewardResponse struct {
	XPGained   int `json:"xp_gained"`
	CoinGained int `json:"coin_gained"`
}

type GameplayProgressionResponse struct {
	LevelBefore      int     `json:"level_before"`
	LevelAfter       int     `json:"level_after"`
	LevelUp          bool    `json:"level_up"`
	XPBefore         int     `json:"xp_before"`
	XPAfter          int     `json:"xp_after"`
	CoinBalanceAfter int     `json:"coin_balance_after"`
	ReputationBefore float64 `json:"reputation_before"`
	ReputationAfter  float64 `json:"reputation_after"`
}

type GameplayFeedbackResponse struct {
	StrengthCategory    string `json:"strength_category"`
	ImprovementCategory string `json:"improvement_category"`
	Message             string `json:"message"`
}

type GameplayIncompleteDetailResponse struct {
	OpenedEvidenceCount   int                               `json:"opened_evidence_count"`
	TotalEvidenceCount    int                               `json:"total_evidence_count"`
	AnsweredQuestionCount int                               `json:"answered_question_count"`
	RequiredQuestionCount int                               `json:"required_question_count"`
	CanTakeDecision       bool                              `json:"can_take_decision"`
	MissingEvidences      []GameplayMissingEvidenceResponse `json:"missing_evidences"`
	MissingQuestions      []GameplayMissingQuestionResponse `json:"missing_questions"`
}

type GameplayMissingEvidenceResponse struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
	Code           string    `json:"code"`
	Label          string    `json:"label"`
	TemplateType   string    `json:"template_type"`
}

type GameplayMissingQuestionResponse struct {
	CaseQuestionID uuid.UUID `json:"case_question_id"`
	Code           string    `json:"code"`
	QuestionType   string    `json:"question_type"`
	QuestionText   string    `json:"question_text"`
}

type GameplaySessionSnapshot struct {
	Case          GameplayCaseSnapshotResponse   `json:"case"`
	ChatbotConfig *GameplayChatbotConfigResponse `json:"chatbot_config,omitempty"`
	Evidences     []GameplayEvidenceResponse     `json:"evidences"`
	Questions     []GameplayQuestionResponse     `json:"questions"`
}

type GameplayChatbotConfigResponse struct {
	BotName               string   `json:"bot_name"`
	BotPersonaDescription string   `json:"bot_persona_description"`
	KnowledgeBoundary     string   `json:"knowledge_boundary"`
	ProhibitedBehaviors   []string `json:"prohibited_behaviors"`
	SuggestedQuestions    []string `json:"suggested_questions"`
}

type GameplayCaseSnapshotResponse struct {
	CaseID                   uuid.UUID  `json:"case_id"`
	CaseVersionID            uuid.UUID  `json:"case_version_id"`
	VersionNumber            int        `json:"version_number"`
	Title                    string     `json:"title"`
	Slug                     string     `json:"slug"`
	ShortDescription         string     `json:"short_description"`
	DifficultyLevel          string     `json:"difficulty_level"`
	RiskLevel                string     `json:"risk_level"`
	EstimatedDurationMinutes int        `json:"estimated_duration_minutes"`
	MinimumLevel             int        `json:"minimum_level"`
	MinimumReputation        float64    `json:"minimum_reputation"`
	ThumbnailURL             *string    `json:"thumbnail_url"`
	PublishedAt              *time.Time `json:"published_at"`
}

type GameplayEvidenceResponse struct {
	CaseEvidenceID     uuid.UUID                               `json:"case_evidence_id"`
	CaseVersionID      uuid.UUID                               `json:"case_version_id"`
	Code               string                                  `json:"code"`
	TemplateType       string                                  `json:"template_type"`
	Label              string                                  `json:"label"`
	SortOrder          int                                     `json:"sort_order"`
	Opened             bool                                    `json:"opened"`
	SocialPost         *GameplaySocialPostEvidenceResponse     `json:"social_post,omitempty"`
	Article            *GameplayArticleEvidenceResponse        `json:"article,omitempty"`
	Blog               *GameplayBlogEvidenceResponse           `json:"blog,omitempty"`
	ForumThread        *GameplayForumThreadEvidenceResponse    `json:"forum_thread,omitempty"`
	ChatTranscript     *GameplayChatTranscriptEvidenceResponse `json:"chat_transcript,omitempty"`
	PublicAnnouncement *GameplayPublicAnnouncementResponse     `json:"public_announcement,omitempty"`
}

type GameplaySocialPostEvidenceResponse struct {
	AuthorName        string    `json:"author_name"`
	AuthorHandle      string    `json:"author_handle"`
	Platform          string    `json:"platform"`
	PostText          string    `json:"post_text"`
	Timestamp         time.Time `json:"timestamp"`
	LikesCount        int       `json:"likes_count"`
	SharesCount       int       `json:"shares_count"`
	CommentsCount     int       `json:"comments_count"`
	IsVerifiedAccount bool      `json:"is_verified_account"`
	ImageURL          *string   `json:"image_url"`
}

type GameplayArticleEvidenceResponse struct {
	Headline    string    `json:"headline"`
	SourceName  string    `json:"source_name"`
	AuthorName  string    `json:"author_name"`
	PublishDate time.Time `json:"publish_date"`
	URL         *string   `json:"url"`
	BodyText    string    `json:"body_text"`
	ImageURL    *string   `json:"image_url"`
}

type GameplayBlogEvidenceResponse struct {
	Title       string    `json:"title"`
	AuthorName  string    `json:"author_name"`
	BlogName    string    `json:"blog_name"`
	PublishDate time.Time `json:"publish_date"`
	BodyText    string    `json:"body_text"`
}

type GameplayForumThreadEvidenceResponse struct {
	ThreadTitle string                            `json:"thread_title"`
	ForumName   string                            `json:"forum_name"`
	Posts       []GameplayForumThreadPostResponse `json:"posts"`
}

type GameplayForumThreadPostResponse struct {
	AuthorName  string    `json:"author_name"`
	Text        string    `json:"text"`
	Timestamp   time.Time `json:"timestamp"`
	UpvoteCount int       `json:"upvote_count"`
	SortOrder   int       `json:"sort_order"`
}

type GameplayChatTranscriptEvidenceResponse struct {
	Participants []GameplayChatTranscriptParticipantResponse `json:"participants"`
	Messages     []GameplayChatTranscriptMessageResponse     `json:"messages"`
}

type GameplayChatTranscriptParticipantResponse struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type GameplayChatTranscriptMessageResponse struct {
	Sender    string    `json:"sender"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	SortOrder int       `json:"sort_order"`
}

type GameplayPublicAnnouncementResponse struct {
	IssuingBody string    `json:"issuing_body"`
	Title       string    `json:"title"`
	BodyText    string    `json:"body_text"`
	Date        time.Time `json:"date"`
}

type GameplayQuestionResponse struct {
	CaseQuestionID      uuid.UUID                         `json:"case_question_id"`
	CaseVersionID       uuid.UUID                         `json:"case_version_id"`
	Code                string                            `json:"code"`
	QuestionType        string                            `json:"question_type"`
	QuestionText        string                            `json:"question_text"`
	IsRequired          bool                              `json:"is_required"`
	SortOrder           int                               `json:"sort_order"`
	Options             []GameplayMCQOptionResponse       `json:"options,omitempty"`
	ConfidenceSlider    *GameplayConfidenceSliderResponse `json:"confidence_slider,omitempty"`
	ClaimClassification []string                          `json:"claim_classification,omitempty"`
	RelatedEvidenceIDs  []uuid.UUID                       `json:"related_evidence_ids"`
}

type GameplayMCQOptionResponse struct {
	OptionCode string `json:"option_code"`
	OptionText string `json:"option_text"`
	SortOrder  int    `json:"sort_order"`
}

type GameplayConfidenceSliderResponse struct {
	MinValue                 int    `json:"min_value"`
	MaxValue                 int    `json:"max_value"`
	SnapInterval             int    `json:"snap_interval"`
	DefaultValue             int    `json:"default_value"`
	LabelLow                 string `json:"label_low"`
	LabelHigh                string `json:"label_high"`
	ShowWarningOnLargeChange bool   `json:"show_warning_on_large_change"`
}
