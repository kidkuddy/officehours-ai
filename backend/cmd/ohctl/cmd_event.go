package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func runEventAdd(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	kind, _ := cmd.Flags().GetString("kind")
	payload, _ := cmd.Flags().GetString("payload")
	if user == "" || kind == "" {
		return fmt.Errorf("--user and --kind are required")
	}
	if !json.Valid([]byte(payload)) {
		return fmt.Errorf("--payload must be valid JSON")
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
		`insert into events (user_id, kind, payload) values ($1,$2,$3) returning id`,
		user, kind, []byte(payload)).Scan(&id)
	if err != nil {
		return err
	}
	printJSON(map[string]any{"id": id, "user_id": user, "kind": kind, "payload": json.RawMessage(payload)})
	return nil
}
