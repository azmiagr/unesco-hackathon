package service

import (
	"github.com/azmiagr/unesco-hackathon/model"
)

func (s *CaseService) GetCaseLookups() (*model.AdminCaseLookupsResponse, error) {
	return &model.AdminCaseLookupsResponse{
		Themes: []model.CaseLookupOptionResponse{
			{Value: model.CaseThemeMisleadingHealthAdvice, Label: "Saran kesehatan menyesatkan"},
			{Value: model.CaseThemeChatbotHallucination, Label: "Halusinasi chatbot"},
			{Value: model.CaseThemeClickbaitHeadline, Label: "Judul artikel manipulatif"},
			{Value: model.CaseThemeStatisticOutOfContext, Label: "Statistik di luar konteks"},
			{Value: model.CaseThemeForumMisinformation, Label: "Validasi informasi keliru di forum"},
			{Value: model.CaseThemeViralConflictContent, Label: "Konten viral yang memperkuat konflik"},
			{Value: model.CaseThemeAlgorithmicEchoChamber, Label: "Sistem rekomendasi/ruang gema"},
			{Value: model.CaseThemeOther, Label: "Lainnya"},
		},
		CompetencyFocuses: []model.CaseLookupOptionResponse{
			{Value: model.CaseCompetencyEvidenceEvaluation, Label: "Evaluasi bukti"},
			{Value: model.CaseCompetencyClaimAnalysis, Label: "Analisis klaim"},
			{Value: model.CaseCompetencyConfidenceCalibration, Label: "Kalibrasi keyakinan"},
			{Value: model.CaseCompetencyReasoning, Label: "Penalaran"},
			{Value: model.CaseCompetencySafetyJudgment, Label: "Penilaian keamanan/keputusan"},
		},
		DifficultyLevels: []model.CaseLookupOptionResponse{
			{Value: model.CaseDifficultyLow, Label: "Easy"},
			{Value: model.CaseDifficultyMedium, Label: "Medium"},
			{Value: model.CaseDifficultyHigh, Label: "Hard"},
		},
		RiskLevels: []model.CaseLookupOptionResponse{
			{Value: model.CaseRiskLow, Label: "Low"},
			{Value: model.CaseRiskMedium, Label: "Medium"},
			{Value: model.CaseRiskHigh, Label: "High"},
		},
		GenerationSources: []model.CaseLookupOptionResponse{
			{Value: model.CaseGenerationManual, Label: "Manual"},
			{Value: model.CaseGenerationAIAssisted, Label: "AI-Assisted"},
		},
	}, nil
}
