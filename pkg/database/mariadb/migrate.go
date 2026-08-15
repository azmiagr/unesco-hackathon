package mariadb

import (
	"github.com/azmiagr/unesco-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.Case{},
		&entity.CaseChatbotConfig{},
		&entity.CaseVersion{},
		&entity.CaseScoringOutcomeConfig{},
		&entity.CaseScoringRule{},
		&entity.CaseOutcomeRule{},
		&entity.CaseOutcomeCityImpactSetting{},
		&entity.CaseEvidence{},
		&entity.CaseEvidenceSocialPost{},
		&entity.CaseEvidenceArticle{},
		&entity.CaseEvidenceBlog{},
		&entity.CaseEvidenceForumThread{},
		&entity.CaseEvidenceForumThreadPost{},
		&entity.CaseEvidenceChatTranscript{},
		&entity.CaseEvidenceChatTranscriptParticipant{},
		&entity.CaseEvidenceChatTranscriptMessage{},
		&entity.CaseEvidencePublicAnnouncement{},
		&entity.CaseQuestion{},
		&entity.CaseQuestionMCQOption{},
		&entity.CaseQuestionEvidenceReference{},
		&entity.CaseQuestionOpenEndedDetail{},
		&entity.CaseQuestionConfidenceSliderDetail{},
		&entity.CaseQuestionClaimClassificationDetail{},
		&entity.CaseSession{},
		&entity.CaseSessionAnswer{},
		&entity.CaseSessionEvidenceProgress{},
		&entity.CaseSessionEvent{},
		&entity.CaseSessionIdempotencyKey{},
		&entity.CaseSessionResult{},
		&entity.AdminLoginOtpSession{},
		&entity.OtpCode{},
		&entity.Avatar{},
		&entity.Title{},
		&entity.ItemCategory{},
		&entity.Item{},
		&entity.UserItem{},
		&entity.RedeemType{},
		&entity.RedeemItem{},
		&entity.UserProfile{},
		&entity.RegistrationSession{},
		&entity.GameConfig{},
		&entity.GameLevel{},
		&entity.AuditLog{},
		&entity.RedeemCode{},
		&entity.CityStatistics{},
	)

	if err != nil {
		return err
	}

	if err := dropLegacyCaseVersionEvidenceColumn(db); err != nil {
		return err
	}
	if err := ensureUserItemPurchaseTypeAllowsGrant(db); err != nil {
		return err
	}
	if err := ensureUserItemTitleOwnershipIndex(db); err != nil {
		return err
	}

	return nil
}
