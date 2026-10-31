import { useEffect, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { useApp, useI18n } from "./context";

export const DOCS_URL = "https://sooauth.com/docs";

export function Button({
  children,
  primary,
  danger,
  onClick,
  type = "button",
  disabled,
  small,
}: {
  children: ReactNode;
  primary?: boolean;
  danger?: boolean;
  onClick?: () => void;
  type?: "button" | "submit";
  disabled?: boolean;
  small?: boolean;
}) {
  const cls = ["button", primary && "primary", danger && "danger", small && "small"].filter(Boolean).join(" ");
  return <button type={type} disabled={disabled} onClick={onClick} className={cls}>{children}</button>;
}

export function Card({ title, description, children, className = "", actions }: { title?: string; description?: ReactNode; children?: ReactNode; className?: string; actions?: ReactNode }) {
  return (
    <section className={`card ${className}`}>
      {title && (
        <div className="card-heading">
          <div>
            <h2>{title}</h2>
            {description && <p>{description}</p>}
          </div>
          {actions && <div className="card-actions">{actions}</div>}
        </div>
      )}
      {children}
    </section>
  );
}

export function Field({ label, value, onChange, type = "text", disabled = false, placeholder, hint }: { label: string; value: string | number; onChange: (value: string) => void; type?: string; disabled?: boolean; placeholder?: string; hint?: string }) {
  return (
    <label className="field">
      <span>{label}</span>
      <input disabled={disabled} type={type} value={value} placeholder={placeholder} onChange={(e) => onChange(e.target.value)} />
      {hint && <small className="field-hint">{hint}</small>}
    </label>
  );
}

export function Check({ label, checked, onChange, hint, disabled }: { label: ReactNode; checked: boolean; onChange: (value: boolean) => void; hint?: ReactNode; disabled?: boolean }) {
  return (
    <label className={`check ${disabled ? "disabled" : ""}`}>
      <input type="checkbox" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />
      <span>
        {label}
        {hint && <small className="check-hint">{hint}</small>}
      </span>
    </label>
  );
}

// OptionCards renders a radio group as selectable cards with a title and a
// one-line explanation, used for choices whose effect needs explaining.
export function OptionCards<T extends string>({ name, value, onChange, options }: { name: string; value: T; onChange: (value: T) => void; options: { value: T; title: string; desc: string; badge?: string }[] }) {
  return (
    <div className="option-cards" role="radiogroup">
      {options.map((opt) => (
        <label key={opt.value} className={`option-card ${value === opt.value ? "active" : ""}`}>
          <input type="radio" name={name} value={opt.value} checked={value === opt.value} onChange={() => onChange(opt.value)} />
          <span className="option-title">{opt.title}{opt.badge && <em>{opt.badge}</em>}</span>
          <span className="option-desc">{opt.desc}</span>
        </label>
      ))}
    </div>
  );
}

export function CopyField({ label, value }: { label: string; value: string }) {
  const { tr } = useI18n();
  const { notify } = useApp();
  return (
    <div className="copy-field">
      <span>{label}</span>
      <div>
        <code>{value}</code>
        <Button small onClick={() => navigator.clipboard.writeText(value).then(() => notify(tr("common.copied")))}>{tr("common.copy")}</Button>
      </div>
    </div>
  );
}

export function Metric({ value, label }: { value: number | string; label: string }) {
  return <div className="metric"><strong>{value}</strong><span>{label}</span></div>;
}

export function Table({ headers, rows }: { headers: string[]; rows: ReactNode[][] }) {
  return (
    <div className="table-wrap">
      <table>
        <thead><tr>{headers.map((header, i) => <th key={header || `h${i}`}>{header}</th>)}</tr></thead>
        <tbody>{rows.map((row, i) => <tr key={i}>{row.map((cell, j) => <td key={j}>{typeof cell === "string" || typeof cell === "number" ? String(cell) : cell}</td>)}</tr>)}</tbody>
      </table>
    </div>
  );
}

export function Empty({ text, children }: { text: string; children?: ReactNode }) {
  return <div className="empty">{text}{children && <div className="empty-actions">{children}</div>}</div>;
}

export function Loading() {
  const { tr } = useI18n();
  return <div className="loading">{tr("common.loading")}</div>;
}

// PageIntro is the short explanation at the top of each settings page, with a
// link to the matching documentation page.
export function PageIntro({ text, docs }: { text: string; docs?: string }) {
  const { tr } = useI18n();
  return (
    <div className="page-intro">
      <p>{text}</p>
      {docs && <a href={`${DOCS_URL}/${docs}`} target="_blank" rel="noopener noreferrer">{tr("common.readDocs")} ↗</a>}
    </div>
  );
}

export function Modal({ title, onClose, children, footer, wide }: { title: string; onClose: () => void; children: ReactNode; footer?: ReactNode; wide?: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  const closeRef = useRef(onClose);
  closeRef.current = onClose;
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") closeRef.current(); };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);
  // Focus once when opening; callers pass a new onClose on every render.
  useEffect(() => {
    const root = ref.current;
    (root?.querySelector<HTMLElement>("input, select, textarea") || root?.querySelector<HTMLElement>("button"))?.focus();
  }, []);
  // Portalled so a modal opened from the dark sidebar keeps the page's theme
  // and is not clipped by the sidebar's mobile transform.
  return createPortal(
    <div className="modal-backdrop" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className={`modal ${wide ? "wide" : ""}`} role="dialog" aria-modal="true" aria-label={title} ref={ref}>
        <div className="modal-head">
          <h2>{title}</h2>
          <button type="button" className="modal-close" onClick={onClose} aria-label="Close">×</button>
        </div>
        <div className="modal-body">{children}</div>
        {footer && <div className="modal-foot">{footer}</div>}
      </div>
    </div>,
    document.body,
  );
}

// ConfirmDialog asks before a destructive action. With confirmText the user
// must type that text (e.g. the user's email) to enable the button.
export function ConfirmDialog({ title, body, confirmLabel, confirmText, onConfirm, onClose }: { title: string; body: ReactNode; confirmLabel: string; confirmText?: string; onConfirm: () => Promise<void> | void; onClose: () => void }) {
  const { tr } = useI18n();
  const [typed, setTyped] = useState("");
  const [busy, setBusy] = useState(false);
  const ready = !confirmText || typed.trim() === confirmText;
  const run = async () => {
    setBusy(true);
    try {
      await onConfirm();
      onClose();
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal
      title={title}
      onClose={onClose}
      footer={<><Button onClick={onClose}>{tr("common.cancel")}</Button><Button danger disabled={!ready || busy} onClick={run}>{busy ? tr("common.saving") : confirmLabel}</Button></>}
    >
      <div className="confirm-body">{body}</div>
      {confirmText && (
        <label className="field confirm-field">
          <span>{tr("common.typeToConfirm", { text: confirmText })}</span>
          <input value={typed} onChange={(e) => setTyped(e.target.value)} placeholder={confirmText} autoComplete="off" />
        </label>
      )}
    </Modal>
  );
}

export function useLoad<T>(loader: () => Promise<T>, deps: unknown[] = []) {
  const { notify } = useApp();
  const [data, setData] = useState<T | null>(null);
  const reload = () => loader().then(setData).catch((e) => notify((e as Error).message));
  useEffect(() => { reload(); }, deps); // eslint-disable-line react-hooks/exhaustive-deps
  return [data, setData, reload] as const;
}
