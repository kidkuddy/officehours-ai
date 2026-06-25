// Polls GET /jobs/:id until it reaches a terminal state, then resolves.
// Returns a runner you call with a job_id; reports status for the UI.
import { useCallback, useRef, useState } from "react";
import { api, type JobResult } from "../api/client";

export type JobPhase = "idle" | "running" | "done" | "error";

const TERMINAL = new Set(["done", "error"]);
const POLL_MS = 1200;

export function useJobPoll() {
  const [phase, setPhase] = useState<JobPhase>("idle");
  const [error, setError] = useState<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const reset = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = null;
    setPhase("idle");
    setError(null);
  }, []);

  // Resolves with the final JobResult, or rejects if the job errors.
  const poll = useCallback((jobId: string): Promise<JobResult> => {
    setPhase("running");
    setError(null);
    return new Promise<JobResult>((resolve, reject) => {
      const tick = async () => {
        try {
          const res = await api.getJob(jobId);
          if (TERMINAL.has(res.status)) {
            if (res.status === "error") {
              setPhase("error");
              setError(res.error || "The agent failed to complete this task.");
              reject(new Error(res.error || "job error"));
              return;
            }
            setPhase("done");
            resolve(res);
            return;
          }
          timer.current = setTimeout(tick, POLL_MS);
        } catch (e) {
          setPhase("error");
          const msg = e instanceof Error ? e.message : "Failed to poll job.";
          setError(msg);
          reject(e);
        }
      };
      void tick();
    });
  }, []);

  return { phase, error, poll, reset };
}
