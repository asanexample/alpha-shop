// Small, on-brand loading / error / empty blocks so a failed or pending fetch never blanks a page.
import type { ReactNode } from "react";

export function LoadingBlock({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="state" role="status" aria-live="polite">
      <div className="state__title">{label}</div>
      <p className="state__body mono">Pulling the latest from the workshop.</p>
    </div>
  );
}

export function ErrorBlock({
  error,
  onRetry,
}: {
  error: unknown;
  onRetry?: () => void;
}) {
  const message =
    error instanceof Error ? error.message : "Something went wrong loading this page.";
  return (
    <div className="state state--error" role="alert">
      <div className="state__title">We couldn't load that</div>
      <p className="state__body">{message}</p>
      {onRetry ? (
        <p style={{ marginTop: "1rem" }}>
          <button type="button" className="btn btn--ghost" onClick={onRetry}>
            Try again
          </button>
        </p>
      ) : null}
    </div>
  );
}

export function EmptyBlock({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="state">
      <div className="state__title">{title}</div>
      {children ? <p className="state__body">{children}</p> : null}
    </div>
  );
}
