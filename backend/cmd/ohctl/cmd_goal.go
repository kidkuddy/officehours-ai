package main

import (
	"fmt"

	"officehours/internal/models"

	"github.com/spf13/cobra"
)

func runGoalCreate(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	title, _ := cmd.Flags().GetString("title")
	desc, _ := cmd.Flags().GetString("desc")
	session, _ := cmd.Flags().GetString("session")
	if user == "" || title == "" {
		return fmt.Errorf("--user and --title are required")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	var id string
	err = pool.QueryRow(ctx,
		`insert into goals (user_id, title, description, source_session_id)
		 values ($1,$2,$3,$4) returning id`,
		user, title, desc, ptrOrNil(session)).Scan(&id)
	if err != nil {
		return err
	}
	printJSON(map[string]any{
		"id": id, "user_id": user, "title": title, "description": desc,
		"source_session_id": ptrOrNil(session), "status": "open",
	})
	return nil
}

func runGoalDone(cmd *cobra.Command, _ []string) error {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		return fmt.Errorf("--id is required")
	}
	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	ct, err := pool.Exec(ctx,
		`update goals set status='done', done_at=now() where id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("goal %s not found", id)
	}
	printJSON(map[string]any{"ok": true, "id": id, "status": "done"})
	return nil
}

func runGoalList(cmd *cobra.Command, _ []string) error {
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

	rows, err := pool.Query(ctx,
		`select id, user_id, title, description, status, source_session_id, created_at, done_at
		 from goals where user_id=$1 order by created_at`, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	out := []models.Goal{}
	for rows.Next() {
		var g models.Goal
		if err := rows.Scan(&g.ID, &g.UserID, &g.Title, &g.Description, &g.Status, &g.SourceSessionID, &g.CreatedAt, &g.DoneAt); err != nil {
			return err
		}
		out = append(out, g)
	}
	printJSON(map[string]any{"goals": out})
	return nil
}
