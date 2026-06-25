package main

import (
	"encoding/json"
	"fmt"

	"officehours/internal/models"

	"github.com/spf13/cobra"
)

func runSignalSet(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.Flags().GetString("user")
	name, _ := cmd.Flags().GetString("name")
	score, _ := cmd.Flags().GetFloat64("score")
	subscores, _ := cmd.Flags().GetString("subscores")
	rationale, _ := cmd.Flags().GetString("rationale")
	floor, _ := cmd.Flags().GetBool("floor")

	if user == "" || name == "" {
		return fmt.Errorf("--user and --name are required")
	}
	if !validSignal(name) {
		return fmt.Errorf("invalid signal name %q (must be one of: %v)", name, models.SignalNames)
	}
	if score < 0 || score > 5 {
		return fmt.Errorf("--score must be 0.0..5.0")
	}
	if !json.Valid([]byte(subscores)) {
		return fmt.Errorf("--subscores must be valid JSON")
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
		`insert into signals (user_id, name, score, subscores, rationale, floor_triggered, updated_at)
		 values ($1,$2,$3,$4,$5,$6, now())
		 on conflict (user_id, name) do update set
		   score=excluded.score, subscores=excluded.subscores,
		   rationale=excluded.rationale, floor_triggered=excluded.floor_triggered,
		   updated_at=now()
		 returning id`,
		user, name, score, []byte(subscores), rationale, floor).Scan(&id)
	if err != nil {
		return err
	}
	printJSON(map[string]any{
		"id": id, "user_id": user, "name": name, "score": score,
		"subscores": json.RawMessage(subscores), "rationale": rationale,
		"floor_triggered": floor,
	})
	return nil
}

func runSignalList(cmd *cobra.Command, _ []string) error {
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
		`select id, user_id, name, score, subscores, rationale, floor_triggered, updated_at
		 from signals where user_id=$1 order by name`, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var s models.Signal
		var sub []byte
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Score, &sub, &s.Rationale, &s.FloorTriggered, &s.UpdatedAt); err != nil {
			return err
		}
		out = append(out, map[string]any{
			"id": s.ID, "user_id": s.UserID, "name": s.Name, "score": s.Score,
			"subscores": json.RawMessage(sub), "rationale": s.Rationale,
			"floor_triggered": s.FloorTriggered, "updated_at": s.UpdatedAt,
		})
	}
	printJSON(map[string]any{"signals": out})
	return nil
}

func validSignal(s string) bool {
	for _, v := range models.SignalNames {
		if v == s {
			return true
		}
	}
	return false
}
