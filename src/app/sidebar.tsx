import { messages } from "@/locales";

/**
 * Projects and their sessions. Empty until session discovery lands; the note
 * explains why the space is blank instead of leaving a void.
 */
export function Sidebar() {
  return (
    <aside className="flex min-h-0 flex-col border-r border-hairline bg-paper">
      <div className="px-4 pt-3 pb-2">
        <h2 className="font-mono text-[11px] tracking-[0.08em] text-ink-3 uppercase">
          {messages.sidebarTitle()}
        </h2>
      </div>
      <p className="px-4 text-[13px] leading-relaxed text-ink-3">{messages.sidebarEmpty()}</p>
    </aside>
  );
}