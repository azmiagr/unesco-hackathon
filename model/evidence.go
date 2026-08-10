package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type GetCaseEvidenceParam struct {
	CaseEvidenceID uuid.UUID
	CaseVersionID  uuid.UUID
	TemplateType   string
}

type ListCaseEvidencesParam struct {
	CaseVersionID uuid.UUID
	TemplateType  string
}

type ReorderCaseEvidenceParam struct {
	CaseEvidenceID uuid.UUID
	SortOrder      int
}

type AdminCreateSocialPostEvidenceRequest struct {
	Label             string                `form:"label" binding:"required"`
	CredibilityTags   string                `form:"credibility_tags" binding:"required"`
	IsCritical        bool                  `form:"is_critical"`
	SortOrder         int                   `form:"sort_order"`
	AuthorName        string                `form:"author_name" binding:"required"`
	AuthorHandle      string                `form:"author_handle" binding:"required"`
	Platform          string                `form:"platform" binding:"required"`
	PostText          string                `form:"post_text" binding:"required"`
	Timestamp         string                `form:"timestamp" binding:"required"`
	LikesCount        int                   `form:"likes_count"`
	SharesCount       int                   `form:"shares_count"`
	CommentsCount     int                   `form:"comments_count"`
	IsVerifiedAccount bool                  `form:"is_verified_account"`
	ImagePrompt       string                `form:"image_prompt"`
	Image             *multipart.FileHeader `form:"image"`
}

type AdminSocialPostEvidenceResponse struct {
	CaseEvidenceID    uuid.UUID `json:"case_evidence_id"`
	CaseVersionID     uuid.UUID `json:"case_version_id"`
	TemplateType      string    `json:"template_type"`
	Label             string    `json:"label"`
	CredibilityTags   []string  `json:"credibility_tags"`
	IsCritical        bool      `json:"is_critical"`
	SortOrder         int       `json:"sort_order"`
	AuthorName        string    `json:"author_name"`
	AuthorHandle      string    `json:"author_handle"`
	Platform          string    `json:"platform"`
	PostText          string    `json:"post_text"`
	Timestamp         time.Time `json:"timestamp"`
	LikesCount        int       `json:"likes_count"`
	SharesCount       int       `json:"shares_count"`
	CommentsCount     int       `json:"comments_count"`
	IsVerifiedAccount bool      `json:"is_verified_account"`
	ImagePrompt       *string   `json:"image_prompt"`
	ImageURL          *string   `json:"image_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AdminCreateSocialPostEvidenceResponse struct {
	Evidence AdminSocialPostEvidenceResponse `json:"evidence"`
}

type AdminUpdateSocialPostEvidenceRequest = AdminCreateSocialPostEvidenceRequest

type AdminUpdateSocialPostEvidenceResponse = AdminCreateSocialPostEvidenceResponse

type AdminCreateArticleEvidenceRequest struct {
	Label           string                `form:"label" binding:"required"`
	CredibilityTags string                `form:"credibility_tags" binding:"required"`
	IsCritical      bool                  `form:"is_critical"`
	SortOrder       int                   `form:"sort_order"`
	Headline        string                `form:"headline" binding:"required"`
	SourceName      string                `form:"source_name" binding:"required"`
	AuthorName      string                `form:"author_name" binding:"required"`
	PublishDate     string                `form:"publish_date" binding:"required"`
	URL             string                `form:"url"`
	BodyText        string                `form:"body_text" binding:"required"`
	ImagePrompt     string                `form:"image_prompt"`
	Image           *multipart.FileHeader `form:"image"`
}

type AdminArticleEvidenceResponse struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID `json:"case_version_id"`
	TemplateType    string    `json:"template_type"`
	Label           string    `json:"label"`
	CredibilityTags []string  `json:"credibility_tags"`
	IsCritical      bool      `json:"is_critical"`
	SortOrder       int       `json:"sort_order"`
	Headline        string    `json:"headline"`
	SourceName      string    `json:"source_name"`
	AuthorName      string    `json:"author_name"`
	PublishDate     time.Time `json:"publish_date"`
	URL             *string   `json:"url"`
	BodyText        string    `json:"body_text"`
	ImagePrompt     *string   `json:"image_prompt"`
	ImageURL        *string   `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminCreateArticleEvidenceResponse struct {
	Evidence AdminArticleEvidenceResponse `json:"evidence"`
}

type AdminUpdateArticleEvidenceRequest = AdminCreateArticleEvidenceRequest

type AdminUpdateArticleEvidenceResponse = AdminCreateArticleEvidenceResponse

type AdminCreateBlogEvidenceRequest struct {
	Label           string `form:"label" binding:"required"`
	CredibilityTags string `form:"credibility_tags" binding:"required"`
	IsCritical      bool   `form:"is_critical"`
	SortOrder       int    `form:"sort_order"`
	Title           string `form:"title" binding:"required"`
	AuthorName      string `form:"author_name" binding:"required"`
	BlogName        string `form:"blog_name" binding:"required"`
	PublishDate     string `form:"publish_date" binding:"required"`
	BodyText        string `form:"body_text" binding:"required"`
}

type AdminBlogEvidenceResponse struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID `json:"case_version_id"`
	TemplateType    string    `json:"template_type"`
	Label           string    `json:"label"`
	CredibilityTags []string  `json:"credibility_tags"`
	IsCritical      bool      `json:"is_critical"`
	SortOrder       int       `json:"sort_order"`
	Title           string    `json:"title"`
	AuthorName      string    `json:"author_name"`
	BlogName        string    `json:"blog_name"`
	PublishDate     time.Time `json:"publish_date"`
	BodyText        string    `json:"body_text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminCreateBlogEvidenceResponse struct {
	Evidence AdminBlogEvidenceResponse `json:"evidence"`
}

type AdminUpdateBlogEvidenceRequest = AdminCreateBlogEvidenceRequest

type AdminUpdateBlogEvidenceResponse = AdminCreateBlogEvidenceResponse

type AdminCreateForumThreadEvidenceRequest struct {
	Label           string                              `json:"label" binding:"required"`
	CredibilityTags []string                            `json:"credibility_tags" binding:"required"`
	IsCritical      bool                                `json:"is_critical"`
	SortOrder       int                                 `json:"sort_order"`
	ThreadTitle     string                              `json:"thread_title" binding:"required"`
	ForumName       string                              `json:"forum_name" binding:"required"`
	Posts           []AdminCreateForumThreadPostRequest `json:"posts" binding:"required"`
}

type AdminCreateForumThreadPostRequest struct {
	AuthorName  string `json:"author_name" binding:"required"`
	Text        string `json:"text" binding:"required"`
	Timestamp   string `json:"timestamp" binding:"required"`
	UpvoteCount int    `json:"upvote_count"`
}

type AdminForumThreadPostResponse struct {
	CaseEvidenceForumThreadPostID uuid.UUID `json:"case_evidence_forum_thread_post_id"`
	CaseEvidenceID                uuid.UUID `json:"case_evidence_id"`
	AuthorName                    string    `json:"author_name"`
	Text                          string    `json:"text"`
	Timestamp                     time.Time `json:"timestamp"`
	UpvoteCount                   int       `json:"upvote_count"`
	SortOrder                     int       `json:"sort_order"`
	CreatedAt                     time.Time `json:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at"`
}

type AdminForumThreadEvidenceResponse struct {
	CaseEvidenceID  uuid.UUID                      `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID                      `json:"case_version_id"`
	TemplateType    string                         `json:"template_type"`
	Label           string                         `json:"label"`
	CredibilityTags []string                       `json:"credibility_tags"`
	IsCritical      bool                           `json:"is_critical"`
	SortOrder       int                            `json:"sort_order"`
	ThreadTitle     string                         `json:"thread_title"`
	ForumName       string                         `json:"forum_name"`
	Posts           []AdminForumThreadPostResponse `json:"posts"`
	CreatedAt       time.Time                      `json:"created_at"`
	UpdatedAt       time.Time                      `json:"updated_at"`
}

type AdminCreateForumThreadEvidenceResponse struct {
	Evidence AdminForumThreadEvidenceResponse `json:"evidence"`
}

type AdminUpdateForumThreadEvidenceRequest = AdminCreateForumThreadEvidenceRequest

type AdminUpdateForumThreadEvidenceResponse = AdminCreateForumThreadEvidenceResponse

type AdminCreateChatTranscriptEvidenceRequest struct {
	Label           string                                    `json:"label" binding:"required"`
	CredibilityTags []string                                  `json:"credibility_tags" binding:"required"`
	IsCritical      bool                                      `json:"is_critical"`
	SortOrder       int                                       `json:"sort_order"`
	Participants    []string                                  `json:"participants" binding:"required"`
	Messages        []AdminCreateChatTranscriptMessageRequest `json:"messages" binding:"required"`
}

type AdminCreateChatTranscriptMessageRequest struct {
	Sender    string `json:"sender" binding:"required"`
	Text      string `json:"text" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
}

type AdminChatTranscriptParticipantResponse struct {
	CaseEvidenceChatTranscriptParticipantID uuid.UUID `json:"case_evidence_chat_transcript_participant_id"`
	CaseEvidenceID                          uuid.UUID `json:"case_evidence_id"`
	Name                                    string    `json:"name"`
	SortOrder                               int       `json:"sort_order"`
	CreatedAt                               time.Time `json:"created_at"`
	UpdatedAt                               time.Time `json:"updated_at"`
}

type AdminChatTranscriptMessageResponse struct {
	CaseEvidenceChatTranscriptMessageID uuid.UUID `json:"case_evidence_chat_transcript_message_id"`
	CaseEvidenceID                      uuid.UUID `json:"case_evidence_id"`
	Sender                              string    `json:"sender"`
	Text                                string    `json:"text"`
	Timestamp                           time.Time `json:"timestamp"`
	SortOrder                           int       `json:"sort_order"`
	CreatedAt                           time.Time `json:"created_at"`
	UpdatedAt                           time.Time `json:"updated_at"`
}

type AdminChatTranscriptEvidenceResponse struct {
	CaseEvidenceID  uuid.UUID                                `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID                                `json:"case_version_id"`
	TemplateType    string                                   `json:"template_type"`
	Label           string                                   `json:"label"`
	CredibilityTags []string                                 `json:"credibility_tags"`
	IsCritical      bool                                     `json:"is_critical"`
	SortOrder       int                                      `json:"sort_order"`
	Participants    []AdminChatTranscriptParticipantResponse `json:"participants"`
	Messages        []AdminChatTranscriptMessageResponse     `json:"messages"`
	CreatedAt       time.Time                                `json:"created_at"`
	UpdatedAt       time.Time                                `json:"updated_at"`
}

type AdminCreateChatTranscriptEvidenceResponse struct {
	Evidence AdminChatTranscriptEvidenceResponse `json:"evidence"`
}

type AdminUpdateChatTranscriptEvidenceRequest = AdminCreateChatTranscriptEvidenceRequest

type AdminUpdateChatTranscriptEvidenceResponse = AdminCreateChatTranscriptEvidenceResponse

type AdminCreatePublicAnnouncementEvidenceRequest struct {
	Label           string   `json:"label" binding:"required"`
	CredibilityTags []string `json:"credibility_tags" binding:"required"`
	IsCritical      bool     `json:"is_critical"`
	SortOrder       int      `json:"sort_order"`
	IssuingBody     string   `json:"issuing_body" binding:"required"`
	Title           string   `json:"title" binding:"required"`
	Date            string   `json:"date" binding:"required"`
	BodyText        string   `json:"body_text" binding:"required"`
}

type AdminPublicAnnouncementEvidenceResponse struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID `json:"case_version_id"`
	TemplateType    string    `json:"template_type"`
	Label           string    `json:"label"`
	CredibilityTags []string  `json:"credibility_tags"`
	IsCritical      bool      `json:"is_critical"`
	SortOrder       int       `json:"sort_order"`
	IssuingBody     string    `json:"issuing_body"`
	Title           string    `json:"title"`
	Date            time.Time `json:"date"`
	BodyText        string    `json:"body_text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminCreatePublicAnnouncementEvidenceResponse struct {
	Evidence AdminPublicAnnouncementEvidenceResponse `json:"evidence"`
}

type AdminUpdatePublicAnnouncementEvidenceRequest = AdminCreatePublicAnnouncementEvidenceRequest

type AdminUpdatePublicAnnouncementEvidenceResponse = AdminCreatePublicAnnouncementEvidenceResponse

type AdminEvidenceDetailResponse struct {
	TemplateType       string                                   `json:"template_type"`
	SocialPost         *AdminSocialPostEvidenceResponse         `json:"social_post,omitempty"`
	Article            *AdminArticleEvidenceResponse            `json:"article,omitempty"`
	Blog               *AdminBlogEvidenceResponse               `json:"blog,omitempty"`
	ForumThread        *AdminForumThreadEvidenceResponse        `json:"forum_thread,omitempty"`
	ChatTranscript     *AdminChatTranscriptEvidenceResponse     `json:"chat_transcript,omitempty"`
	PublicAnnouncement *AdminPublicAnnouncementEvidenceResponse `json:"public_announcement,omitempty"`
}

type AdminDeleteCaseEvidenceResponse struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
}
