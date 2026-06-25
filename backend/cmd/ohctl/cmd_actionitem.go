package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runActionItemCreate(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	session, _ := cmd.Flags().GetString("session")
	title, _ := cmd.Flags().GetString("title")
	horizon, _ := cmd.Flags().GetString("horizon")
	rationale, _ := cmd.Flags().GetString("rationale")
	programRef, _ := cmd.Flags().GetString("program-ref")
	if user == "" || title == "" {
		return fmt.Errorf("--user and --title are required")
	}
	switch horizon {
	case "immediate", "short", "medium":
	default:
		return fmt.Errorf("--horizon must be immediate|short|medium")
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
		`insert into action_items (user_id, session_id, title, horizon, rationale, program_ref)
		 values ($1,$2,$3,$4,$5,$6) returning id`,
		user, ptrOrNil(session), title, horizon, rationale, programRef).Scan(&id)
	if err != nil {
		return err
	}
	printJSON(map[string]any{
		"id": id, "user_id": user, "session_id": ptrOrNil(session),
		"title": title, "horizon": horizon, "rationale": rationale,
		"program_ref": programRef, "status": "open",
	})
	return nil
}
