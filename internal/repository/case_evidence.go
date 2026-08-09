package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseEvidenceRepository interface {
	CreateSocialPostEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, socialPost *entity.CaseEvidenceSocialPost) error
	CreateArticleEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, article *entity.CaseEvidenceArticle) error
	CreateBlogEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, blog *entity.CaseEvidenceBlog) error
	CreateForumThreadEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, forumThread *entity.CaseEvidenceForumThread, posts []entity.CaseEvidenceForumThreadPost) error
	CreateChatTranscriptEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, chatTranscript *entity.CaseEvidenceChatTranscript, participants []entity.CaseEvidenceChatTranscriptParticipant, messages []entity.CaseEvidenceChatTranscriptMessage) error
	CreatePublicAnnouncementEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, publicAnnouncement *entity.CaseEvidencePublicAnnouncement) error
	GetCaseEvidence(tx *gorm.DB, param model.GetCaseEvidenceParam) (*entity.CaseEvidence, error)
	GetCaseEvidenceForUpdate(tx *gorm.DB, evidenceID uuid.UUID) (*entity.CaseEvidence, error)
	ListCaseEvidences(tx *gorm.DB, param model.ListCaseEvidencesParam) ([]entity.CaseEvidence, error)
	ListAdminCaseEvidenceRows(tx *gorm.DB, caseVersionID uuid.UUID) ([]model.AdminCaseEvidenceListRow, error)
	UpdateCaseEvidence(tx *gorm.DB, evidence *entity.CaseEvidence) error
	UpdateSocialPostEvidence(tx *gorm.DB, socialPost *entity.CaseEvidenceSocialPost) error
	UpdateArticleEvidence(tx *gorm.DB, article *entity.CaseEvidenceArticle) error
	UpdateBlogEvidence(tx *gorm.DB, blog *entity.CaseEvidenceBlog) error
	UpdateForumThreadEvidence(tx *gorm.DB, forumThread *entity.CaseEvidenceForumThread) error
	ReplaceForumThreadPosts(tx *gorm.DB, caseEvidenceID uuid.UUID, posts []entity.CaseEvidenceForumThreadPost) error
	UpdateChatTranscriptEvidence(tx *gorm.DB, chatTranscript *entity.CaseEvidenceChatTranscript) error
	ReplaceChatTranscriptParticipants(tx *gorm.DB, caseEvidenceID uuid.UUID, participants []entity.CaseEvidenceChatTranscriptParticipant) error
	ReplaceChatTranscriptMessages(tx *gorm.DB, caseEvidenceID uuid.UUID, messages []entity.CaseEvidenceChatTranscriptMessage) error
	UpdatePublicAnnouncementEvidence(tx *gorm.DB, publicAnnouncement *entity.CaseEvidencePublicAnnouncement) error
	DeleteCaseEvidence(tx *gorm.DB, evidenceID uuid.UUID) error
	UpdateCaseEvidenceSortOrder(tx *gorm.DB, param model.ReorderCaseEvidenceParam) error
}

type CaseEvidenceRepository struct {
	db *gorm.DB
}

func NewCaseEvidenceRepository(db *gorm.DB) ICaseEvidenceRepository {
	return &CaseEvidenceRepository{db: db}
}

func (r *CaseEvidenceRepository) CreateSocialPostEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, socialPost *entity.CaseEvidenceSocialPost) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(socialPost).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) CreateArticleEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, article *entity.CaseEvidenceArticle) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(article).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) CreateBlogEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, blog *entity.CaseEvidenceBlog) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(blog).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) CreateForumThreadEvidence(
	tx *gorm.DB,
	evidence *entity.CaseEvidence,
	forumThread *entity.CaseEvidenceForumThread,
	posts []entity.CaseEvidenceForumThreadPost,
) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(forumThread).Error
	if err != nil {
		return err
	}

	if len(posts) > 0 {
		err := tx.Debug().Create(&posts).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseEvidenceRepository) CreateChatTranscriptEvidence(
	tx *gorm.DB,
	evidence *entity.CaseEvidence,
	chatTranscript *entity.CaseEvidenceChatTranscript,
	participants []entity.CaseEvidenceChatTranscriptParticipant,
	messages []entity.CaseEvidenceChatTranscriptMessage,
) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(chatTranscript).Error
	if err != nil {
		return err
	}

	if len(participants) > 0 {
		err = tx.Debug().Create(&participants).Error
		if err != nil {
			return err
		}
	}

	if len(messages) > 0 {
		err = tx.Debug().Create(&messages).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseEvidenceRepository) CreatePublicAnnouncementEvidence(
	tx *gorm.DB,
	evidence *entity.CaseEvidence,
	publicAnnouncement *entity.CaseEvidencePublicAnnouncement,
) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(publicAnnouncement).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) GetCaseEvidence(tx *gorm.DB, param model.GetCaseEvidenceParam) (*entity.CaseEvidence, error) {
	var evidence entity.CaseEvidence
	query := preloadEvidenceDetails(tx)

	if param.CaseEvidenceID != uuid.Nil {
		query = query.Where("case_evidence_id = ?", param.CaseEvidenceID)
	}
	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.TemplateType != "" {
		query = query.Where("template_type = ?", param.TemplateType)
	}

	if err := query.First(&evidence).Error; err != nil {
		return nil, err
	}

	return &evidence, nil
}

func (r *CaseEvidenceRepository) GetCaseEvidenceForUpdate(tx *gorm.DB, evidenceID uuid.UUID) (*entity.CaseEvidence, error) {
	var evidence entity.CaseEvidence

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_evidence_id = ?", evidenceID).
		First(&evidence).Error
	if err != nil {
		return nil, err
	}

	return &evidence, nil
}

func (r *CaseEvidenceRepository) ListCaseEvidences(tx *gorm.DB, param model.ListCaseEvidencesParam) ([]entity.CaseEvidence, error) {
	var evidences []entity.CaseEvidence
	query := preloadEvidenceDetails(tx)

	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.TemplateType != "" {
		query = query.Where("template_type = ?", param.TemplateType)
	}

	err := query.
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&evidences).Error
	if err != nil {
		return nil, err
	}

	return evidences, nil
}

func (r *CaseEvidenceRepository) ListAdminCaseEvidenceRows(tx *gorm.DB, caseVersionID uuid.UUID) ([]model.AdminCaseEvidenceListRow, error) {
	var evidences []model.AdminCaseEvidenceListRow

	err := tx.Model(&entity.CaseEvidence{}).
		Select([]string{
			"case_evidences.case_evidence_id",
			"case_evidences.case_version_id",
			"case_evidences.template_type",
			"case_evidences.label",
			"case_evidences.is_critical",
			`CASE
				WHEN social_posts.image_url IS NOT NULL AND social_posts.image_url <> '' THEN true
				WHEN articles.image_url IS NOT NULL AND articles.image_url <> '' THEN true
				ELSE false
			END AS has_image`,
			"case_evidences.sort_order",
			"case_evidences.created_at",
			"case_evidences.updated_at",
		}).
		Joins(`
			LEFT JOIN case_evidence_social_posts social_posts
				ON social_posts.case_evidence_id = case_evidences.case_evidence_id
		`).
		Joins(`
			LEFT JOIN case_evidence_articles articles
				ON articles.case_evidence_id = case_evidences.case_evidence_id
		`).
		Where("case_evidences.case_version_id = ?", caseVersionID).
		Order("case_evidences.sort_order ASC").
		Order("case_evidences.created_at ASC").
		Scan(&evidences).Error
	if err != nil {
		return nil, err
	}

	return evidences, nil
}

func (r *CaseEvidenceRepository) UpdateCaseEvidence(tx *gorm.DB, evidence *entity.CaseEvidence) error {
	err := tx.Debug().Save(evidence).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) UpdateSocialPostEvidence(tx *gorm.DB, socialPost *entity.CaseEvidenceSocialPost) error {
	err := tx.Debug().Save(socialPost).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) UpdateArticleEvidence(tx *gorm.DB, article *entity.CaseEvidenceArticle) error {
	err := tx.Debug().Save(article).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) UpdateBlogEvidence(tx *gorm.DB, blog *entity.CaseEvidenceBlog) error {
	err := tx.Debug().Save(blog).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) UpdateForumThreadEvidence(tx *gorm.DB, forumThread *entity.CaseEvidenceForumThread) error {
	err := tx.Debug().Save(forumThread).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) ReplaceForumThreadPosts(
	tx *gorm.DB,
	caseEvidenceID uuid.UUID,
	posts []entity.CaseEvidenceForumThreadPost,
) error {
	err := tx.Debug().
		Where("case_evidence_id = ?", caseEvidenceID).
		Delete(&entity.CaseEvidenceForumThreadPost{}).Error
	if err != nil {
		return err
	}

	if len(posts) > 0 {
		err = tx.Debug().Create(&posts).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseEvidenceRepository) UpdateChatTranscriptEvidence(tx *gorm.DB, chatTranscript *entity.CaseEvidenceChatTranscript) error {
	err := tx.Debug().Save(chatTranscript).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) ReplaceChatTranscriptParticipants(
	tx *gorm.DB,
	caseEvidenceID uuid.UUID,
	participants []entity.CaseEvidenceChatTranscriptParticipant,
) error {
	err := tx.Debug().
		Where("case_evidence_id = ?", caseEvidenceID).
		Delete(&entity.CaseEvidenceChatTranscriptParticipant{}).Error
	if err != nil {
		return err
	}

	if len(participants) > 0 {
		err = tx.Debug().Create(&participants).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseEvidenceRepository) ReplaceChatTranscriptMessages(
	tx *gorm.DB,
	caseEvidenceID uuid.UUID,
	messages []entity.CaseEvidenceChatTranscriptMessage,
) error {
	err := tx.Debug().
		Where("case_evidence_id = ?", caseEvidenceID).
		Delete(&entity.CaseEvidenceChatTranscriptMessage{}).Error
	if err != nil {
		return err
	}

	if len(messages) > 0 {
		err := tx.Debug().Create(&messages).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseEvidenceRepository) UpdatePublicAnnouncementEvidence(
	tx *gorm.DB,
	publicAnnouncement *entity.CaseEvidencePublicAnnouncement,
) error {
	err := tx.Debug().Save(publicAnnouncement).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) DeleteCaseEvidence(tx *gorm.DB, evidenceID uuid.UUID) error {
	err := tx.Debug().
		Where("case_evidence_id = ?", evidenceID).
		Delete(&entity.CaseEvidence{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) UpdateCaseEvidenceSortOrder(tx *gorm.DB, param model.ReorderCaseEvidenceParam) error {
	err := tx.Debug().
		Model(&entity.CaseEvidence{}).
		Where("case_evidence_id = ?", param.CaseEvidenceID).
		Update("sort_order", param.SortOrder).Error
	if err != nil {
		return err
	}
	return nil
}

func preloadEvidenceDetails(query *gorm.DB) *gorm.DB {
	return query.
		Preload("SocialPost").
		Preload("Article").
		Preload("Blog").
		Preload("ForumThread.Posts", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("ChatTranscript.Participants", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("ChatTranscript.Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("PublicAnnouncement")
}
