import { useCallback, useEffect, useRef, useState } from "react";
import { api, type Document } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState, Spinner, relativeTime } from "../components/ui";
import { IconUpload, IconDoc } from "../components/Icons";

const ACCEPT = [".md", ".txt", ".pdf"];

function fileTag(name: string): string {
  const ext = name.split(".").pop()?.toLowerCase();
  if (ext === "pdf") return "PDF";
  if (ext === "md") return "MD";
  return "TXT";
}

export default function DataRoom() {
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [drag, setDrag] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let active = true;
    api
      .listDocuments()
      .then((d) => active && setDocs(d))
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load your documents."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  const upload = useCallback(async (file: File) => {
    setError(null);
    const ext = "." + (file.name.split(".").pop()?.toLowerCase() ?? "");
    if (!ACCEPT.includes(ext)) {
      setError(`That file type isn't supported. Use ${ACCEPT.join(", ")}.`);
      return;
    }
    setUploading(true);
    try {
      const { document } = await api.uploadDocument(file);
      setDocs((prev) => [document, ...prev]);
    } catch (e) {
      setError(e instanceof Error ? e.message : "The upload didn't finish. Try again.");
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = "";
    }
  }, []);

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDrag(false);
    const f = e.dataTransfer.files?.[0];
    if (f) void upload(f);
  };

  if (loading) return <PageLoading label="Loading your data room" />;

  return (
    <div className="content page-fade-in">
      <div className="page-header">
        <div className="page-eyebrow">Records</div>
        <h1 className="page-title">Data Room</h1>
        <p className="page-subtitle">
          Upload documents to your private collection. Advisors can cite them in sessions.
        </p>
      </div>

      {error && <ErrorAlert message={error} />}

      <div
        className={`dropzone${drag ? " drag" : ""}`}
        onClick={() => !uploading && fileRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDrag(true);
        }}
        onDragLeave={() => setDrag(false)}
        onDrop={onDrop}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") fileRef.current?.click(); }}
        style={{ marginBottom: 8 }}
      >
        <input
          ref={fileRef}
          type="file"
          accept={ACCEPT.join(",")}
          style={{ display: "none" }}
          onChange={(e) => {
            const f = e.target.files?.[0];
            if (f) void upload(f);
          }}
        />
        <div className="dz-icon">{uploading ? <Spinner /> : <IconUpload />}</div>
        <div style={{ fontWeight: 600, fontSize: 15, fontFamily: "var(--font-display)" }}>
          {uploading ? "Indexing your document…" : "Drop a file, or click to upload"}
        </div>
        <div style={{ marginTop: 6, fontSize: 12, color: "var(--slate-2)", fontFamily: "var(--font-mono)", letterSpacing: "0.06em" }}>
          {ACCEPT.join("  ·  ")}
        </div>
      </div>

      <div className="section-label">Documents ({docs.length})</div>
      {docs.length === 0 ? (
        <EmptyState
          icon={<IconDoc />}
          title="No documents yet"
          hint="Upload a pitch deck, financial model, or any supporting doc to ground your sessions."
        />
      ) : (
        <div className="list">
          {docs.map((d) => (
            <div className="list-row" key={d.id}>
              <div className="lr-icon" style={{ fontSize: 10 }}>{fileTag(d.filename)}</div>
              <div className="lr-main">
                <div className="lr-title">{d.filename}</div>
                <div className="lr-sub">{d.collection} · {relativeTime(d.created_at)}</div>
              </div>
              <span className="badge">Indexed</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
