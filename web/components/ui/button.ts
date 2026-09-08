import { cn } from "../../lib/cn";

type Variant = "primary" | "secondary" | "ghost";
type Size = "sm" | "md";

const base =
  "inline-flex cursor-pointer items-center justify-center gap-2 rounded-none font-sans font-medium tracking-tight transition-[background-color,border-color,transform,color] duration-200 focus-visible:outline-2";

const sizes: Record<Size, string> = {
  sm: "h-9 px-4 text-sm",
  md: "h-11 px-6 text-sm",
};

const variants: Record<Variant, string> = {
  primary:
    "bg-accent text-accent-fg hover:bg-accent-hover hover:-translate-y-px active:translate-y-0",
  secondary:
    "border-[1.5px] border-line-strong bg-transparent text-fg hover:bg-bg-subtle hover:-translate-y-px active:translate-y-0",
  ghost:
    "border-[1.5px] border-transparent bg-transparent text-fg-muted hover:bg-bg-subtle hover:text-fg",
};

export function buttonClass(
  variant: Variant = "primary",
  size: Size = "md",
  extra?: string,
) {
  return cn(base, sizes[size], variants[variant], extra);
}
