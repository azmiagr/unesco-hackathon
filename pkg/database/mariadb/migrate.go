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

	return nil
}

func dropLegacyCaseVersionEvidenceColumn(db *gorm.DB) error {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		AND table_name = ?
		AND column_name = ?
	`, "case_versions", "evidence").Scan(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	return db.Exec("ALTER TABLE case_versions DROP COLUMN evidence").Error
}
