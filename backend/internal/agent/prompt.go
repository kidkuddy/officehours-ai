package agent

import (
	"fmt"
	"strings"
)

// PromptContext carries the runtime values injected into a rendered prompt.
type PromptContext struct {
	UserID     string
	SessionID  string
	AdvisorKey string
	// LatestMessage is the founder's most recent message (advisor/learn jobs).
	LatestMessage string
	// Opening marks an opening turn: the advisor starts the conversation with a
	// short first message and there is no founder message yet.
	Opening bool
	// OnboardingText is the company description (diagnoser job).
	OnboardingText string
	// Outcomes is the conclude note (scorer job).
	Outcomes string
}

// stages and signal names are repeated here as plain strings so the agent
// package does not depend on models (which is allowed but kept decoupled).
const stagesLine = "Ideation, Market Validation, Structuration, Fundraising, Launch Planning, Growth"
const signalsLine = "Market, Commercial Offer, Innovation, Scalability, Green"

// RenderAdvisor builds the prompt for an advisor or learn (concept) chat turn.
func RenderAdvisor(def *Definition, c PromptContext) string {
	var b strings.Builder
	b.WriteString(def.Body)
	b.WriteString("\n\n---\n\n")
	b.WriteString("## Runtime context (this turn)\n\n")
	b.WriteString(fmt.Sprintf("- user_uuid: `%s`\n", c.UserID))
	b.WriteString(fmt.Sprintf("- session_uuid: `%s`\n", c.SessionID))
	if def.Collection != "" {
		b.WriteString(fmt.Sprintf("- your KB collection: `%s`\n", def.Collection))
	}
	b.WriteString("\nYou have `ohctl` on your PATH and `DATABASE_URL` is set. ")
	b.WriteString("First read context with `ohctl session get " + c.SessionID + "` and ")
	b.WriteString("`ohctl profile get --user " + c.UserID + "`. ")
	if def.Collection != "" {
		b.WriteString("Ground your answer with `ohctl rag query --collection " + def.Collection + " \"<q>\" --k 5` and cite program names. ")
	}
	if c.Opening || strings.TrimSpace(c.LatestMessage) == "" {
		// Opening turn: the advisor starts the conversation. There is no founder
		// message yet, so produce a SHORT, welcoming first message that opens
		// the session and invites the founder to talk.
		b.WriteString("\n\n## Opening turn (no founder message yet)\n\n")
		b.WriteString("This is the FIRST message of the session — the founder has not written anything yet. ")
		b.WriteString("Open the conversation with a SHORT first message (2-4 sentences max): greet the founder by your advisor persona, ")
		b.WriteString("say in one line what you can help with, and ask one focused opening question to get them talking. ")
		b.WriteString("Do not dump a long checklist. Keep it warm and brief.\n\n")
	} else {
		b.WriteString("\n\n## Founder's latest message\n\n")
		b.WriteString(c.LatestMessage)
		b.WriteString("\n\n")
	}
	b.WriteString("Reply directly with your final assistant message in markdown. ")
	b.WriteString("Do NOT call `ohctl session message` to store the reply — the backend captures your final text as the chat reply. ")
	b.WriteString("You MAY create a goal or action item via `ohctl` if you commit to a concrete next step.")
	return b.String()
}

// RenderDiagnoser builds the prompt for the onboarding diagnoser job.
func RenderDiagnoser(def *Definition, c PromptContext) string {
	var b strings.Builder
	b.WriteString(def.Body)
	b.WriteString("\n\n---\n\n")
	b.WriteString("## Runtime context\n\n")
	b.WriteString(fmt.Sprintf("- user_uuid: `%s`\n", c.UserID))
	b.WriteString("- valid stages (exact strings): " + stagesLine + "\n")
	b.WriteString("- the 5 Signals (exact names): " + signalsLine + "\n\n")
	b.WriteString("You have `ohctl` on your PATH and `DATABASE_URL` is set. ")
	b.WriteString("Read `ohctl profile get --user " + c.UserID + "` and ground stage reasoning with `ohctl rag query`. ")
	b.WriteString("Persist EVERYTHING via `ohctl` — there is no other output channel:\n")
	b.WriteString("- `ohctl profile set-stage --user " + c.UserID + " --stage \"<stage>\" --evidence '[\"...\"]'`\n")
	b.WriteString("- set initial values for all 5 Signals with `ohctl signal set --user " + c.UserID + " --name \"Market\" --score <n> --subscores '<json>' --rationale \"...\"`\n")
	b.WriteString("- append a `ohctl event add --user " + c.UserID + " --kind stage_change --payload '{...}'`\n\n")
	b.WriteString("## Founder's company description (onboarding)\n\n")
	b.WriteString(c.OnboardingText)
	return b.String()
}

// RenderScorer builds the prompt for the scorer job (session conclude).
func RenderScorer(def *Definition, c PromptContext) string {
	var b strings.Builder
	b.WriteString(def.Body)
	b.WriteString("\n\n---\n\n")
	b.WriteString("## Runtime context\n\n")
	b.WriteString(fmt.Sprintf("- user_uuid: `%s`\n", c.UserID))
	b.WriteString(fmt.Sprintf("- session_uuid: `%s`\n", c.SessionID))
	b.WriteString("- valid stages (exact strings): " + stagesLine + "\n")
	b.WriteString("- the 5 Signals (exact names): " + signalsLine + "\n\n")
	b.WriteString("You have `ohctl` on your PATH and `DATABASE_URL` is set. ")
	b.WriteString("Read context with `ohctl session get " + c.SessionID + "`, `ohctl profile get --user " + c.UserID + "`, and `ohctl signal list --user " + c.UserID + "`. ")
	b.WriteString("Follow the methodology in the body exactly and persist EVERYTHING via `ohctl`. ")
	b.WriteString("Finish by concluding the session: `ohctl session conclude " + c.SessionID + " --outcomes \"...\"`.\n")
	if strings.TrimSpace(c.Outcomes) != "" {
		b.WriteString("\nFounder-provided conclude note: " + c.Outcomes + "\n")
	}
	return b.String()
}
