package main

import (
	"encoding/json"
	"fmt"

	"officehours/internal/rag"

	"github.com/spf13/cobra"
)

func runRagIndex(cmd *cobra.Command, args []string) error {
	folder := args[0]
	collection, _ := cmd.Flags().GetString("collection")
	user, _ := cmd.Flags().GetString("user")
	if collection == "" {
		return fmt.Errorf("--collection is required")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	res, err := rag.IndexFolder(ctx, pool, folder, collection, ptrOrNil(user))
	if err != nil {
		return err
	}
	b, _ := json.Marshal(res)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	printJSON(m)
	return nil
}

func runRagQuery(cmd *cobra.Command, args []string) error {
	q := args[0]
	collection, _ := cmd.Flags().GetString("collection")
	k, _ := cmd.Flags().GetInt("k")
	user, _ := cmd.Flags().GetString("user")
	if collection == "" {
		return fmt.Errorf("--collection is required")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	hits, err := rag.Query(ctx, pool, collection, q, k, ptrOrNil(user))
	if err != nil {
		return err
	}
	printJSON(map[string]any{
		"collection": collection,
		"query":      q,
		"k":          k,
		"results":    hits,
	})
	return nil
}

// ptrOrNil returns a pointer to s, or nil if s is empty.
func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
