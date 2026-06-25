package main

import (
	"encoding/json"
	"fmt"

	"officehours/internal/models"

	"github.com/spf13/cobra"
)

func runProfileGet(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	if user == "" {
		return fmt.Errorf("--user is required")
	}
	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	var p models.Profile
	var evidence []byte
	err = pool.QueryRow(ctx,
		`select user_id, company_text, stage, stage_evidence, created_at, updated_at
		 from profiles where user_id=$1`, user).
		Scan(&p.UserID, &p.CompanyText, &p.Stage, &evidence, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("profile not found for user %s: %w", user, err)
	}
	printJSON(map[string]any{
		"user_id":        p.UserID,
		"company_text":   p.CompanyText,
		"stage":          p.Stage,
		"stage_evidence": json.RawMessage(evidence),
		"created_at":     p.CreatedAt,
		"updated_at":     p.UpdatedAt,
	})
	return nil
}

func runProfileSetStage(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	stage, _ := cmd.Flags().GetString("stage")
	evidence, _ := cmd.Flags().GetString("evidence")
	if user == "" || stage == "" {
		return fmt.Errorf("--user and --stage are required")
	}
	if !validStage(stage) {
		return fmt.Errorf("invalid stage %q (must be one of: %v)", stage, models.Stages)
	}
	if !json.Valid([]byte(evidence)) {
		return fmt.Errorf("--evidence must be valid JSON")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	ct, err := pool.Exec(ctx,
		`update profiles set stage=$2, stage_evidence=$3, updated_at=now() where user_id=$1`,
		user, stage, []byte(evidence))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		// Create profile if missing.
		_, err = pool.Exec(ctx,
			`insert into profiles (user_id, stage, stage_evidence) values ($1,$2,$3)`,
			user, stage, []byte(evidence))
		if err != nil {
			return err
		}
	}
	printJSON(map[string]any{"ok": true, "user_id": user, "stage": stage, "stage_evidence": json.RawMessage(evidence)})
	return nil
}

func validStage(s string) bool {
	for _, v := range models.Stages {
		if v == s {
			return true
		}
	}
	return false
}
