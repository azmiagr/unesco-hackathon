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
		&entity.CaseVersion{},
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
		&entity.AdminLoginOtpSession{},
		&entity.OtpCode{},
		&entity.Avatar{},
		&entity.UserProfile{},
		&entity.RegistrationSession{},
	)

	if err != nil {
		return err
	}

	return nil
}
