/**
 * Swiss abstract — CSS only, decorative. Intersecting hairlines over a
 * 12-column module grid with a few filled rectangles in beige / taupe / fg.
 */
export function GridFigure() {
  return (
    <div
      aria-hidden
      className="relative w-full border border-line-strong bg-bg-subtle"
      style={{ aspectRatio: "4 / 5" }}
    >
      {/* module grid */}
      <div className="absolute inset-0 grid grid-cols-6 grid-rows-8">
        {Array.from({ length: 48 }).map((_, i) => (
          <div key={i} className="border-b border-r border-line/60" />
        ))}
      </div>

      {/* filled rectangles */}
      <div className="absolute left-0 top-0 h-[37.5%] w-1/2 bg-taupe" />
      <div className="absolute bottom-0 right-0 h-1/4 w-2/3 bg-fg" />
      <div className="absolute left-1/2 top-1/2 h-[12.5%] w-1/3 -translate-y-1/2 bg-beige" />

      {/* heavy crosshair */}
      <div className="absolute inset-y-0 left-1/2 w-px bg-line-strong" />
      <div className="absolute inset-x-0 top-[37.5%] h-px bg-line-strong" />

      {/* metadata tick */}
      <span className="absolute bottom-2 left-2 font-mono text-[10px] uppercase tracking-[0.14em] text-fg-muted">
        OIDC / 01
      </span>
    </div>
  );
}
