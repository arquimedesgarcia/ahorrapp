package config

import (
	"os"
	"testing"
)

func TestLoyaltyConfig_Defaults(t *testing.T) {
	for _, k := range []string{
		"LOYALTY_BASE_POINTS",
		"LOYALTY_FIRST_OBSERVATION_BONUS",
		"LOYALTY_DATA_COMPLETION_BONUS",
		"LOYALTY_DAILY_AWARD_CAP",
	} {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Cleanup(func() { os.Unsetenv("DATABASE_URL") })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LoyaltyBasePoints != 10 {
		t.Errorf("LoyaltyBasePoints default = %d, want 10", cfg.LoyaltyBasePoints)
	}
	if cfg.LoyaltyFirstObservationBonus != 5 {
		t.Errorf("LoyaltyFirstObservationBonus default = %d, want 5", cfg.LoyaltyFirstObservationBonus)
	}
	if cfg.LoyaltyDataCompletionBonus != 3 {
		t.Errorf("LoyaltyDataCompletionBonus default = %d, want 3", cfg.LoyaltyDataCompletionBonus)
	}
	if cfg.LoyaltyDailyAwardCap != 20 {
		t.Errorf("LoyaltyDailyAwardCap default = %d, want 20", cfg.LoyaltyDailyAwardCap)
	}
}

func TestLoyaltyConfig_OverridesParse(t *testing.T) {
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("LOYALTY_BASE_POINTS", "42")
	t.Setenv("LOYALTY_FIRST_OBSERVATION_BONUS", "7")
	t.Setenv("LOYALTY_DATA_COMPLETION_BONUS", "11")
	t.Setenv("LOYALTY_DAILY_AWARD_CAP", "3")
	t.Cleanup(func() { os.Unsetenv("DATABASE_URL") })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LoyaltyBasePoints != 42 {
		t.Errorf("LoyaltyBasePoints = %d, want 42", cfg.LoyaltyBasePoints)
	}
	if cfg.LoyaltyFirstObservationBonus != 7 {
		t.Errorf("LoyaltyFirstObservationBonus = %d, want 7", cfg.LoyaltyFirstObservationBonus)
	}
	if cfg.LoyaltyDataCompletionBonus != 11 {
		t.Errorf("LoyaltyDataCompletionBonus = %d, want 11", cfg.LoyaltyDataCompletionBonus)
	}
	if cfg.LoyaltyDailyAwardCap != 3 {
		t.Errorf("LoyaltyDailyAwardCap = %d, want 3", cfg.LoyaltyDailyAwardCap)
	}
}
