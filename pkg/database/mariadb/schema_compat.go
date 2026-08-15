package mariadb

import "gorm.io/gorm"

func ensureUserItemPurchaseTypeAllowsGrant(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		AND table_name = ?
		AND column_name = ?
	`, "user_items", "purchase_type").Scan(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	return db.Exec("ALTER TABLE user_items MODIFY purchase_type enum('shop','redeem','grant') NOT NULL DEFAULT 'shop'").Error
}

func ensureUserItemTitleOwnershipIndex(db *gorm.DB) error {
	type indexColumn struct {
		ColumnName string `gorm:"column:column_name"`
	}

	var columns []indexColumn
	if err := db.Raw(`
		SELECT column_name
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		AND table_name = ?
		AND index_name = ?
		ORDER BY seq_in_index ASC
	`, "user_items", "uq_user_items_user_title").Scan(&columns).Error; err != nil {
		return err
	}

	if len(columns) == 2 && columns[0].ColumnName == "user_id" && columns[1].ColumnName == "title_id" {
		return nil
	}
	if len(columns) > 0 {
		if err := db.Exec("ALTER TABLE user_items DROP INDEX uq_user_items_user_title").Error; err != nil {
			return err
		}
	}

	return db.Exec("CREATE UNIQUE INDEX uq_user_items_user_title ON user_items (user_id, title_id)").Error
}

func dropLegacyCaseVersionEvidenceColumn(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		AND table_name = ?
		AND column_name = ?
	`, "case_versions", "evidence").Scan(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	return db.Exec("ALTER TABLE case_versions DROP COLUMN evidence").Error
}
