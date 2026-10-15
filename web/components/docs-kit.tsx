import type { ReactNode } from "react";
import Link from "next/link";

// Small building blocks shared by the documentation pages so every guide has
// the same rhythm: title, lead, sections, code, tables and notes.

export function DocTitle({ children, lead }: { children: ReactNode; lead?: ReactNode }) {
  return (
    <header>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">{children}</h1>
      {lead && <p className="mt-3 max-w-[64ch] text-[16px] leading-[1.6] text-fg-muted">{lead}</p>}
    </header>
  );
}

export function H2({ id, children }: { id: string; children: ReactNode }) {
  return (
    <h2 id={id} className="mt-12 scroll-mt-24 font-sans text-[22px] font-semibold text-fg">
      <a href={`#${id}`} className="hover:underline underline-offset-4">{children}</a>
    </h2>
  );
}

export function H3({ id, children }: { id?: string; children: ReactNode }) {
  return <h3 id={id} className="mt-8 scroll-mt-24 font-sans text-[17px] font-semibold text-fg">{children}</h3>;
}

export function P({ children }: { children: ReactNode }) {
  return <p className="mt-3 max-w-[70ch] text-[15px] leading-[1.65] text-fg-muted">{children}</p>;
}

export function C({ children }: { children: ReactNode }) {
  return <code className="rounded-sm bg-bg-subtle px-1.5 py-0.5 font-mono text-[13px] text-fg">{children}</code>;
}

export function Code({ children, title }: { children: string; title?: string }) {
  return (
    <figure className="mt-4">
      {title && <figcaption className="font-mono text-xs text-fg-muted mb-1.5">{title}</figcaption>}
      <pre className="overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{children}</pre>
    </figure>
  );
}

export function Table({ head, rows }: { head: string[]; rows: ReactNode[][] }) {
  return (
    <div className="mt-4 overflow-x-auto border border-line">
      <table className="w-full text-left text-[14px]">
        <thead className="bg-bg-subtle">
          <tr>{head.map((h) => <th key={h} className="px-3 py-2 font-sans text-xs font-semibold uppercase tracking-wide text-fg-muted">{h}</th>)}</tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-t border-line align-top">
              {row.map((cell, j) => <td key={j} className="px-3 py-2.5 leading-[1.55] text-fg-muted">{cell}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function Note({ children, tone = "info" }: { children: ReactNode; tone?: "info" | "warn" }) {
  return (
    <div className={`mt-5 border-l-2 px-4 py-3 text-[14px] leading-[1.6] ${tone === "warn" ? "border-accent bg-accent/5 text-fg" : "border-line bg-bg-subtle text-fg-muted"}`}>
      {children}
    </div>
  );
}

export function List({ children }: { children: ReactNode }) {
  return <ul className="mt-3 grid max-w-[70ch] list-disc gap-1.5 pl-5 text-[15px] leading-[1.6] text-fg-muted">{children}</ul>;
}

export function Steps({ children }: { children: ReactNode }) {
  return <ol className="mt-3 grid max-w-[70ch] list-decimal gap-1.5 pl-5 text-[15px] leading-[1.6] text-fg-muted">{children}</ol>;
}

export function DocLink({ href, children }: { href: string; children: ReactNode }) {
  const cls = "text-fg underline underline-offset-2";
  return href.startsWith("/") ? <Link href={href} className={cls}>{children}</Link> : <a href={href} className={cls}>{children}</a>;
}

export function DashboardPath({ children }: { children: ReactNode }) {
  return <span className="font-medium text-fg">{children}</span>;
}
