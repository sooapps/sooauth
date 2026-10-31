import { useEffect, useMemo, useRef, useState } from "react";
import type { Data } from "../api";
import { useI18n } from "../context";
import { Icon } from "../icons";
import { accountSections, sections } from "../nav";
import { accountPath, homePath, navigate, projectPath } from "../router";

type Entry = { key: string; label: string; hint: string; icon: string; run: () => void };

// CommandPalette (⌘K) jumps to any project or page by typing part of its name.
export function CommandPalette({ me, onClose }: { me: Data | null; onClose: () => void }) {
  const { tr } = useI18n();
  const [query, setQuery] = useState("");
  const [cursor, setCursor] = useState(0);
  const input = useRef<HTMLInputElement>(null);
  useEffect(() => input.current?.focus(), []);

  const entries = useMemo<Entry[]>(() => {
    const activeId: string = me?.project?.id || "";
    const projects: Data[] = me?.projects || [];
    return [
      { key: "home", label: tr("nav.dashboard"), hint: "", icon: "home", run: () => navigate(homePath) },
      ...sections.map((s) => ({ key: s.id, label: tr(s.labelKey), hint: me?.project?.name || "", icon: s.icon, run: () => navigate(projectPath(activeId, s.id)) })),
      ...projects.map((p) => ({ key: `p:${p.id}`, label: p.name, hint: tr("nav.projects"), icon: "folder", run: () => navigate(projectPath(p.id, "overview")) })),
      ...accountSections.map((s) => ({ key: s.id, label: tr(s.labelKey), hint: tr("nav.account"), icon: s.icon, run: () => navigate(accountPath(s.id)) })),
    ];
  }, [me, tr]);

  const q = query.trim().toLocaleLowerCase();
  const results = q ? entries.filter((e) => e.label.toLocaleLowerCase().includes(q)) : entries;
  const pick = (entry?: Entry) => {
    if (!entry) return;
    entry.run();
    onClose();
  };
  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") { e.preventDefault(); setCursor((cursor + 1) % Math.max(results.length, 1)); }
    if (e.key === "ArrowUp") { e.preventDefault(); setCursor((cursor - 1 + results.length) % Math.max(results.length, 1)); }
    if (e.key === "Enter") pick(results[cursor]);
    if (e.key === "Escape") onClose();
  };

  return (
    <div className="palette-backdrop" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="palette" role="dialog" aria-modal="true" aria-label={tr("search.placeholder")}>
        <div className="palette-input">
          <Icon name="search" />
          <input ref={input} value={query} placeholder={tr("search.placeholder")} onKeyDown={onKey} onChange={(e) => { setQuery(e.target.value); setCursor(0); }} />
        </div>
        <ul>
          {results.map((entry, i) => (
            <li key={entry.key}>
              <button type="button" className={i === cursor ? "active" : ""} onMouseEnter={() => setCursor(i)} onClick={() => pick(entry)}>
                <Icon name={entry.icon} />
                <span>{entry.label}</span>
                {entry.hint && <small>{entry.hint}</small>}
              </button>
            </li>
          ))}
          {!results.length && <li className="palette-empty">{tr("search.empty")}</li>}
        </ul>
      </div>
    </div>
  );
}
