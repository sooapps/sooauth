import type { ReactElement } from "react";

// Stroke icons for the dashboard shell, drawn on a 24px grid.
const paths: Record<string, ReactElement> = {
  home: <><path d="M4 11.5 12 5l8 6.5" /><path d="M6 10v9h12v-9" /><path d="M10 19v-5h4v5" /></>,
  folder: <path d="M4 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2z" />,
  rocket: <><path d="M12 15 9 12c.8-3.4 3-6.6 8-8 1.2 5-1 9-5 11Z" /><path d="M9 12H5l2-3h3" /><path d="M12 15v4l3-2v-3" /><circle cx="14" cy="9" r="1" /></>,
  gauge: <><path d="M4 16a8 8 0 1 1 16 0" /><path d="m12 16 3.5-5" /><path d="M3 19h18" /></>,
  key: <><circle cx="8" cy="15" r="4" /><path d="m11 12 8-8" /><path d="m16 7 3 3" /></>,
  users: <><circle cx="9" cy="9" r="3.5" /><path d="M3 20a6 6 0 0 1 12 0" /><path d="M16 5.8a3.5 3.5 0 0 1 0 6.4" /><path d="M18 14.5a6 6 0 0 1 3 5.5" /></>,
  form: <><rect x="4" y="4" width="16" height="16" rx="2" /><path d="M8 9h8" /><path d="M8 13h8" /><path d="M8 17h4" /></>,
  shield: <><path d="M12 3 5 6v5c0 4.5 2.8 8 7 10 4.2-2 7-5.5 7-10V6z" /><path d="m9 12 2 2 4-4" /></>,
  plug: <><path d="M9 3v6" /><path d="M15 3v6" /><path d="M6 9h12v3a6 6 0 0 1-12 0z" /><path d="M12 18v3" /></>,
  sparkle: <><path d="M12 4c.5 4 2 5.5 6 6-4 .5-5.5 2-6 6-.5-4-2-5.5-6-6 4-.5 5.5-2 6-6Z" /><path d="M19 15c.2 1.7.8 2.3 2 2.5-1.2.2-1.8.8-2 2.5-.2-1.7-.8-2.3-2-2.5 1.2-.2 1.8-.8 2-2.5Z" /></>,
  play: <><circle cx="12" cy="12" r="9" /><path d="m10 8.5 5 3.5-5 3.5z" /></>,
  hook: <><circle cx="7" cy="17" r="3" /><circle cx="17" cy="17" r="3" /><circle cx="12" cy="7" r="3" /><path d="m10.5 9.6-2.2 4.8" /><path d="m13.5 9.6 2.2 4.8" /><path d="M10 17h4" /></>,
  devices: <><rect x="3" y="5" width="13" height="10" rx="2" /><rect x="15" y="10" width="6" height="9" rx="1.5" /><path d="M6 19h6" /></>,
  history: <><path d="M4 12a8 8 0 1 0 2.5-5.8L4 8.5" /><path d="M4 4v4.5h4.5" /><path d="M12 8v4l3 2" /></>,
  palette: <><path d="M12 3a9 9 0 1 0 0 18c1.5 0 2-1 2-2s-.5-1.5-.5-2.5 1-2 2-2H18a3 3 0 0 0 3-3c0-5-4-8.5-9-8.5Z" /><circle cx="8" cy="11" r="1" /><circle cx="11" cy="7.5" r="1" /><circle cx="15.5" cy="8.5" r="1" /></>,
  mail: <><rect x="3" y="5" width="18" height="14" rx="2" /><path d="m4 7 8 6 8-6" /></>,
  gear: <><circle cx="12" cy="12" r="3" /><path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M18.4 5.6l-2.1 2.1M7.7 16.3l-2.1 2.1" /></>,
  card: <><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M3 10h18" /><path d="M7 15h3" /></>,
  account: <><circle cx="12" cy="9" r="4" /><path d="M4.5 20a8 8 0 0 1 15 0" /></>,
  book: <><path d="M5 5a2 2 0 0 1 2-2h12v15H7a2 2 0 0 0-2 2z" /><path d="M5 20a2 2 0 0 0 2 1h12v-3" /></>,
  flask: <><path d="M10 3v6L5 18a2 2 0 0 0 1.8 3h10.4A2 2 0 0 0 19 18l-5-9V3" /><path d="M8.5 3h7" /><path d="M8 15h8" /></>,
  plus: <><path d="M12 5v14" /><path d="M5 12h14" /></>,
  back: <path d="m15 5-7 7 7 7" />,
  next: <path d="m9 5 7 7-7 7" />,
  search: <><circle cx="11" cy="11" r="7" /><path d="m20 20-3.8-3.8" /></>,
  panel: <><rect x="3" y="4" width="18" height="16" rx="3" /><path d="M9 4v16" /></>,
  menu: <><path d="M4 7h16" /><path d="M4 12h16" /><path d="M4 17h16" /></>,
  updown: <><path d="m8 9 4-4 4 4" /><path d="m8 15 4 4 4-4" /></>,
  external: <><path d="M14 5h5v5" /><path d="m19 5-8 8" /><path d="M18 14v4a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4" /></>,
};

export type IconName = string;

export function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {paths[name]}
    </svg>
  );
}
