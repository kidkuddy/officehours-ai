package main

import (
	"fmt"

	"officehours/internal/models"

	"github.com/spf13/cobra"
)

func runSessionGet(_ *cobra.Command, args []string) error {
	id := args[0]
	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	var s models.Session
	err = pool.QueryRow(ctx,
		`select id, user_id, kind, advisor_key, title, status, outcomes, created_at, concluded_at
		 from sessions where id=$1`, id).
		Scan(&s.ID, &s.UserID, &s.Kind, &s.AdvisorKey, &s.Title, &s.Status, &s.Outcomes, &s.CreatedAt, &s.ConcludedAt)
	if err != nil {
		return fmt.Errorf("session %s not found: %w", id, err)
	}

	msgRows, err := pool.Query(ctx,
		`select id, session_id, role, content, created_at
		 from messages where session_id=$1 order by created_at`, id)
	if err != nil {
		return err
	}
	defer msgRows.Close()
	messages := []models.Message{}
	for msgRows.Next() {
		var m models.Message
		if err := msgRows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return err
		}
		messages = append(messages, m)
	}

	aiRows, err := pool.Query(ctx,
		`select id, user_id, session_id, title, horizon, rationale, program_ref, status, created_at
		 from action_items where session_id=$1 order by created_at`, id)
	if err != nil {
		return err
	}
	defer aiRows.Close()
	items := []models.ActionItem{}
	for aiRows.Next() {
		var a models.ActionItem
		if err := aiRows.Scan(&a.ID, &a.UserID, &a.SessionID, &a.Title, &a.Horizon, &a.Rationale, &a.ProgramRef, &a.Status, &a.CreatedAt); err != nil {
			return err
		}
		items = append(items, a)
	}

	printJSON(map[string]any{
		"session":      s,
		"messages":     messages,
		"action_items": items,
	})
	return nil
}

func runSessionMessage(cmd *cobra.Command, args []string) error {
	id := args[0]
	role, _ := cmd.Flags().GetString("role")
	content, _ := cmd.Flags().GetString("content")
	switch role {
	case "user", "assistant", "system":
	default:
		return fmt.Errorf("--role must be user|assistant|system")
	}
	if content == "" {
		return fmt.Errorf("--content is required")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	var msgID string
	err = pool.QueryRow(ctx,
		`insert into messages (session_id, role, content) values ($1,$2,$3) returning id`,
		id, role, content).Scan(&msgID)
	if err != nil {
		return err
	}
	printJSON(map[string]any{"id": msgID, "session_id": id, "role": role, "content": content})
	return nil
}

func runSessionConclude(cmd *cobra.Command, args []string) error {
	id := args[0]
	outcomes, _ := cmd.Flags().GetString("outcomes")

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	ct, err := pool.Exec(ctx,
		`update sessions set status='concluded', outcomes=$2, concluded_at=now() where id=$1`,
		id, outcomes)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("session %s not found", id)
	}
	printJSON(map[string]any{"ok": true, "id": id, "status": "concluded", "outcomes": outcomes})
	return nil
}
