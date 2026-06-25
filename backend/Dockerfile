# syntax=docker/dockerfile:1

# --- build stage: compile api + ohctl ---
FROM golang:1.25-bookworm AS build
WORKDIR /src

# Cache deps
COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api  ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/ohctl ./cmd/ohctl

# --- final stage: node base so the `claude` CLI is available ---
# We run on a Node base and install @anthropic-ai/claude-code globally so the
# `claude` binary is on PATH. The job worker execs `claude -p ...` per task.
FROM node:20-bookworm-slim AS final

# Install the Claude Code CLI (`claude`) and the Gemini CLI (`gemini`) on PATH.
# The job worker execs one or the other depending on AGENT_PROVIDER.
RUN npm install -g @anthropic-ai/claude-code @google/gemini-cli \
 && npm cache clean --force

# gosu lets the entrypoint drop from root to the non-root `node` user.
RUN apt-get update \
 && apt-get install -y --no-install-recommends gosu \
 && rm -rf /var/lib/apt/lists/*

# Copy compiled Go binaries; put ohctl on PATH for the agent.
COPY --from=build /out/api   /usr/local/bin/api
COPY --from=build /out/ohctl /usr/local/bin/ohctl

# Bundle migrations so the api can apply schema on boot if it chooses to.
COPY migrations /app/migrations

ENV UPLOAD_DIR=/data/uploads \
    SEED_DIR=/seed \
    CONFIG_DIR=/config \
    PORT=8080 \
    HOME=/home/node
RUN mkdir -p /data/uploads

# AUTH (Plan A): mount the host's logged-in Claude subscription credentials at
# /host-claude (read-only). The entrypoint copies them into /home/node/.claude
# (the non-root user's home) because `claude` refuses --dangerously-skip-permissions
# under root. Run `claude` login on the host BEFORE bringing the stack up.
# See docker-compose.yml: ${HOME}/.claude:/host-claude:ro
# If ANTHROPIC_API_KEY is set, the CLI uses that instead.
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["api"]
