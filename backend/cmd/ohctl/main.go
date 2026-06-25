// Command ohctl is the agent's CLI. It connects to DATABASE_URL and every
// command prints JSON to stdout. See BUILD_SPEC §3 for the frozen surface.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		// Errors are printed as JSON to stdout for the agent; exit non-zero.
		printJSON(map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ohctl",
		Short:         "OfficeHours.ai agent CLI (talks directly to Postgres; prints JSON)",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newSeedCmd(),
		newRagCmd(),
		newProfileCmd(),
		newSignalCmd(),
		newGoalCmd(),
		newActionItemCmd(),
		newSessionCmd(),
		newEventCmd(),
	)
	return root
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "ohctl: encode error: %v\n", err)
	}
}

// --- seed ---

func newSeedCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "seed", Short: "Seed users and demo data"}

	user := &cobra.Command{
		Use:   "user",
		Short: "Create a login",
		RunE:  runSeedUser,
	}
	user.Flags().String("email", "", "user email")
	user.Flags().String("password", "", "user password")
	user.Flags().String("name", "", "user name")

	demo := &cobra.Command{
		Use:   "demo",
		Short: "Seed advisors+learn from /seed md, index /seed/kb, optional demo user",
		RunE:  runSeedDemo,
	}
	demo.Flags().Bool("with-user", false, "also create a demo user")

	cmd.AddCommand(user, demo)
	return cmd
}

// --- rag ---

func newRagCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "rag", Short: "RAG indexing and query"}

	index := &cobra.Command{
		Use:   "index <folder>",
		Short: "Chunk md/txt/pdf, store + tsv",
		Args:  cobra.ExactArgs(1),
		RunE:  runRagIndex,
	}
	index.Flags().String("collection", "", "collection name")
	index.Flags().String("user", "", "user uuid (optional)")

	query := &cobra.Command{
		Use:   "query \"<q>\"",
		Short: "Full-text query a collection",
		Args:  cobra.ExactArgs(1),
		RunE:  runRagQuery,
	}
	query.Flags().String("collection", "", "collection name")
	query.Flags().Int("k", 5, "number of results")
	query.Flags().String("user", "", "user uuid (optional)")

	cmd.AddCommand(index, query)
	return cmd
}

// --- profile ---

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Profile get/set"}

	get := &cobra.Command{Use: "get", Short: "Get the profile", RunE: runProfileGet}
	get.Flags().String("user", "", "user uuid")

	setStage := &cobra.Command{Use: "set-stage", Short: "Set maturity stage with evidence", RunE: runProfileSetStage}
	setStage.Flags().String("user", "", "user uuid")
	setStage.Flags().String("stage", "", "one of the 6 stages")
	setStage.Flags().String("evidence", "[]", "json array of evidence")

	cmd.AddCommand(get, setStage)
	return cmd
}

// --- signal ---

func newSignalCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "signal", Short: "Signal set/list"}

	set := &cobra.Command{Use: "set", Short: "Set a composite signal", RunE: runSignalSet}
	set.Flags().String("user", "", "user uuid")
	set.Flags().String("name", "", "signal name (one of the 5)")
	set.Flags().Float64("score", 0, "0.0..5.0")
	set.Flags().String("subscores", "[]", "json array of subscores")
	set.Flags().String("rationale", "", "rationale text")
	set.Flags().Bool("floor", false, "floor triggered")

	list := &cobra.Command{Use: "list", Short: "List signals", RunE: runSignalList}
	list.Flags().String("user", "", "user uuid")

	cmd.AddCommand(set, list)
	return cmd
}

// --- goal ---

func newGoalCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "goal", Short: "Goal create/done/list"}

	create := &cobra.Command{Use: "create", Short: "Create a goal", RunE: runGoalCreate}
	create.Flags().String("user", "", "user uuid")
	create.Flags().String("title", "", "goal title")
	create.Flags().String("desc", "", "goal description")
	create.Flags().String("session", "", "source session uuid")

	done := &cobra.Command{Use: "done", Short: "Mark a goal done", RunE: runGoalDone}
	done.Flags().String("id", "", "goal uuid")

	list := &cobra.Command{Use: "list", Short: "List goals", RunE: runGoalList}
	list.Flags().String("user", "", "user uuid")

	cmd.AddCommand(create, done, list)
	return cmd
}

// --- action-item ---

func newActionItemCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "action-item", Short: "Action item commands"}

	create := &cobra.Command{Use: "create", Short: "Create an action item", RunE: runActionItemCreate}
	create.Flags().String("user", "", "user uuid")
	create.Flags().String("session", "", "session uuid")
	create.Flags().String("title", "", "title")
	create.Flags().String("horizon", "short", "immediate|short|medium")
	create.Flags().String("rationale", "", "rationale")
	create.Flags().String("program-ref", "", "KB source for grounding")

	cmd.AddCommand(create)
	return cmd
}

// --- session ---

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "session", Short: "Session get/message/conclude"}

	get := &cobra.Command{
		Use:   "get <uuid>",
		Short: "Meta + messages + action items",
		Args:  cobra.ExactArgs(1),
		RunE:  runSessionGet,
	}

	message := &cobra.Command{
		Use:   "message <uuid>",
		Short: "Append a message to a session",
		Args:  cobra.ExactArgs(1),
		RunE:  runSessionMessage,
	}
	message.Flags().String("role", "assistant", "user|assistant|system")
	message.Flags().String("content", "", "message content")

	conclude := &cobra.Command{
		Use:   "conclude <uuid>",
		Short: "Conclude a session",
		Args:  cobra.ExactArgs(1),
		RunE:  runSessionConclude,
	}
	conclude.Flags().String("outcomes", "", "outcomes text")

	cmd.AddCommand(get, message, conclude)
	return cmd
}

// --- event ---

func newEventCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "event", Short: "Mon Parcours timeline events"}

	add := &cobra.Command{Use: "add", Short: "Add a timeline event", RunE: runEventAdd}
	add.Flags().String("user", "", "user uuid")
	add.Flags().String("kind", "", "event kind")
	add.Flags().String("payload", "{}", "json payload")

	cmd.AddCommand(add)
	return cmd
}
