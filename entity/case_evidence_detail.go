package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseEvidenceSocialPost struct {
	CaseEvidenceID    uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	AuthorName        string    `json:"author_name" gorm:"type:varchar(150);not null"`
	AuthorHandle      string    `json:"author_handle" gorm:"type:varchar(100);not null"`
	Platform          string    `json:"platform" gorm:"type:varchar(100);not null"`
	PostText          string    `json:"post_text" gorm:"type:text;not null"`
	Timestamp         time.Time `json:"timestamp" gorm:"not null"`
	LikesCount        int       `json:"likes_count" gorm:"type:int;not null;default:0"`
	SharesCount       int       `json:"shares_count" gorm:"type:int;not null;default:0"`
	CommentsCount     int       `json:"comments_count" gorm:"type:int;not null;default:0"`
	IsVerifiedAccount bool      `json:"is_verified_account" gorm:"not null;default:false"`
	ImagePrompt       *string   `json:"image_prompt" gorm:"type:text"`
	ImageURL          *string   `json:"image_url" gorm:"type:varchar(500)"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidenceArticle struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	Headline       string    `json:"headline" gorm:"type:varchar(250);not null"`
	SourceName     string    `json:"source_name" gorm:"type:varchar(150);not null"`
	AuthorName     string    `json:"author_name" gorm:"type:varchar(150);not null"`
	PublishDate    time.Time `json:"publish_date" gorm:"type:date;not null"`
	URL            *string   `json:"url" gorm:"type:varchar(500)"`
	BodyText       string    `json:"body_text" gorm:"type:longtext;not null"`
	ImagePrompt    *string   `json:"image_prompt" gorm:"type:text"`
	ImageURL       *string   `json:"image_url" gorm:"type:varchar(500)"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidenceBlog struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	Title          string    `json:"title" gorm:"type:varchar(250);not null"`
	AuthorName     string    `json:"author_name" gorm:"type:varchar(150);not null"`
	BlogName       string    `json:"blog_name" gorm:"type:varchar(150);not null"`
	PublishDate    time.Time `json:"publish_date" gorm:"type:date;not null"`
	BodyText       string    `json:"body_text" gorm:"type:longtext;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidenceForumThread struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	ThreadTitle    string    `json:"thread_title" gorm:"type:varchar(250);not null"`
	ForumName      string    `json:"forum_name" gorm:"type:varchar(150);not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Posts []CaseEvidenceForumThreadPost `json:"posts" gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
}

type CaseEvidenceForumThreadPost struct {
	CaseEvidenceForumThreadPostID uuid.UUID `json:"case_evidence_forum_thread_post_id" gorm:"type:varchar(36);primaryKey"`
	CaseEvidenceID                uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);not null;index"`
	AuthorName                    string    `json:"author_name" gorm:"type:varchar(150);not null"`
	Text                          string    `json:"text" gorm:"type:text;not null"`
	Timestamp                     time.Time `json:"timestamp" gorm:"not null"`
	UpvoteCount                   int       `json:"upvote_count" gorm:"type:int;not null;default:0"`
	SortOrder                     int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt                     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidenceChatTranscript struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Participants []CaseEvidenceChatTranscriptParticipant `json:"participants" gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	Messages     []CaseEvidenceChatTranscriptMessage     `json:"messages" gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
}

type CaseEvidenceChatTranscriptParticipant struct {
	CaseEvidenceChatTranscriptParticipantID uuid.UUID `json:"case_evidence_chat_transcript_participant_id" gorm:"type:varchar(36);primaryKey"`
	CaseEvidenceID                          uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);not null;index"`
	Name                                    string    `json:"name" gorm:"type:varchar(150);not null"`
	SortOrder                               int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt                               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidenceChatTranscriptMessage struct {
	CaseEvidenceChatTranscriptMessageID uuid.UUID `json:"case_evidence_chat_transcript_message_id" gorm:"type:varchar(36);primaryKey"`
	CaseEvidenceID                      uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);not null;index"`
	Sender                              string    `json:"sender" gorm:"type:varchar(150);not null"`
	Text                                string    `json:"text" gorm:"type:text;not null"`
	Timestamp                           time.Time `json:"timestamp" gorm:"not null"`
	SortOrder                           int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt                           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseEvidencePublicAnnouncement struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	IssuingBody    string    `json:"issuing_body" gorm:"type:varchar(200);not null"`
	Title          string    `json:"title" gorm:"type:varchar(250);not null"`
	BodyText       string    `json:"body_text" gorm:"type:longtext;not null"`
	Date           time.Time `json:"date" gorm:"type:date;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
