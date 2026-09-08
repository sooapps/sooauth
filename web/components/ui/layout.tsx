import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

export function Container({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return <div className={cn("container-swiss", className)}>{children}</div>;
}

export function Section({
  children,
  className,
  id,
  labelledBy,
  divider = true,
}: {
  children: ReactNode;
  className?: string;
  id?: string;
  labelledBy?: string;
  divider?: boolean;
}) {
  return (
    <section
      id={id}
      aria-labelledby={labelledBy}
      className={cn(
        "scroll-mt-24 py-[clamp(4rem,8vw,8rem)]",
        divider && "border-t border-line",
        className,
      )}
    >
      <Container>{children}</Container>
    </section>
  );
}

export function Eyebrow({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <p
      className={cn(
        "font-mono text-xs uppercase tracking-[0.14em] text-fg-muted",
        className,
      )}
    >
      {children}
    </p>
  );
}

export function SectionHeading({
  eyebrow,
  title,
  lead,
  id,
}: {
  eyebrow?: string;
  title: string;
  lead?: string;
  id?: string;
}) {
  return (
    <header className="max-w-[52ch]">
      {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
      <h2
        id={id}
        className={cn(
          "font-sans text-[clamp(1.75rem,3vw,2.25rem)] font-bold leading-[1.1] tracking-[-0.02em] text-fg",
          eyebrow && "mt-4",
        )}
      >
        {title}
      </h2>
      {lead ? (
        <p className="mt-4 text-[16px] leading-[1.6] text-fg-muted">{lead}</p>
      ) : null}
    </header>
  );
}

export function Card({
  children,
  className,
  as: Tag = "div",
}: {
  children: ReactNode;
  className?: string;
  as?: "div" | "article" | "li";
}) {
  return (
    <Tag
      className={cn(
        "rounded-none border border-line bg-bg-subtle p-6 shadow-card md:p-8",
        className,
      )}
    >
      {children}
    </Tag>
  );
}
