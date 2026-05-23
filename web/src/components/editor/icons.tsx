export function Icon({ path, viewBox = '0 0 24 24', strokeWidth = 2 }: { path: string; viewBox?: string; strokeWidth?: number }) {
  return (
    <svg className="size-4" viewBox={viewBox} fill="none" stroke="currentColor" strokeWidth={strokeWidth}>
      <path d={path} />
    </svg>
  )
}

function IconSvg({ children, viewBox = '0 0 24 24', strokeWidth = 2 }: { children: React.ReactNode; viewBox?: string; strokeWidth?: number }) {
  return (
    <svg className="size-4" viewBox={viewBox} fill="none" stroke="currentColor" strokeWidth={strokeWidth}>
      {children}
    </svg>
  )
}

export function Undo() {
  return <Icon path="M3 7v6h6 M21 17a9 9 0 00-9-9 9 9 0 00-6 2.3L3 13" />
}

export function Redo() {
  return <Icon path="M21 7v6h-6 M3 17a9 9 0 019-9 9 9 0 016 2.3l3 2.7" />
}

export function Bold() {
  return <Icon path="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" />
}

export function Italic() {
  return <Icon path="M19 4h-9 M14 20H5 M15 4L9 20" />
}

export function Underline() {
  return <Icon path="M6 3v7a6 6 0 0 0 6 6 6 6 0 0 0 6-6V3 M4 21h16" />
}

export function Strikethrough() {
  return <Icon path="M17.3 4.9c-2.3-.6-4.4-1-6.2-.9-2.7 0-5.3.7-5.3 3.6 0 1.5 1.1 2.6 3.3 3.4 M4 12h16 M17.7 13.3c.5.8.7 1.7.7 2.6 0 3.5-3.4 4.1-6.4 4.1-2.3 0-4.5-.4-6.5-1" />
}

export function SuperscriptIcon() {
  return <span className="text-xs font-semibold leading-none">x²</span>
}

export function SubscriptIcon() {
  return <span className="text-xs font-semibold leading-none">x₂</span>
}

export function Color() {
  return <Icon path="M15.243 4.515l-6.738 6.737-.707 2.121-1.04 1.041 2.828 2.829 1.04-1.041 2.122-.707 6.737-6.738-4.242-4.242zm6.364 3.536a1 1 0 010 1.414l-7.778 7.778-2.122.707-1.414 1.414a1 1 0 01-1.414 0l-4.243-4.243a1 1 0 010-1.414l1.414-1.414.707-2.121 7.778-7.778a1 1 0 011.414 0l5.657 5.657z" />
}

export function Link() {
  return <Icon path="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
}

export function Image() {
  return <Icon path="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M17 8l-5-5-5 5 M12 3v12" />
}

export function Heading1() {
  return <span className="text-xs font-semibold">H1</span>
}

export function Heading2() {
  return <span className="text-xs font-semibold">H2</span>
}

export function Heading3() {
  return <span className="text-xs font-semibold">H3</span>
}

export function BulletList() {
  return (
    <IconSvg>
      <line x1="9" y1="6" x2="20" y2="6" /><line x1="9" y1="12" x2="20" y2="12" /><line x1="9" y1="18" x2="20" y2="18" />
      <circle cx="5" cy="6" r="1" fill="currentColor" /><circle cx="5" cy="12" r="1" fill="currentColor" /><circle cx="5" cy="18" r="1" fill="currentColor" />
    </IconSvg>
  )
}

export function OrderedList() {
  return (
    <IconSvg>
      <line x1="10" y1="6" x2="21" y2="6" /><line x1="10" y1="12" x2="21" y2="12" /><line x1="10" y1="18" x2="21" y2="18" />
      <text x="3" y="8" fontSize="8" fill="currentColor" stroke="none">1</text>
      <text x="3" y="14" fontSize="8" fill="currentColor" stroke="none">2</text>
      <text x="3" y="20" fontSize="8" fill="currentColor" stroke="none">3</text>
    </IconSvg>
  )
}

export function TaskList() {
  return (
    <IconSvg>
      <rect x="3" y="3" width="18" height="18" rx="2" />
      <path d="M9 12l2 2 4-4" />
    </IconSvg>
  )
}

export function Blockquote() {
  return <Icon path="M4.583 17.321C3.553 16.227 3 15 3 13.011c0-3.5 2.457-6.637 6.03-8.188l.893 1.378c-3.335 1.804-3.987 4.145-4.247 5.621.537-.278 1.24-.375 1.929-.311 1.804.167 3.226 1.648 3.226 3.489a3.5 3.5 0 01-3.5 3.5c-1.073 0-2.099-.49-2.748-1.179zm10 0C13.553 16.227 13 15 13 13.011c0-3.5 2.457-6.637 6.03-8.188l.893 1.378c-3.335 1.804-3.987 4.145-4.247 5.621.537-.278 1.24-.375 1.929-.311 1.804.167 3.226 1.648 3.226 3.489a3.5 3.5 0 01-3.5 3.5c-1.073 0-2.099-.49-2.748-1.179z" />
}

export function Code() {
  return <Icon path="M16 18l6-6-6-6 M8 6l-6 6 6 6" />
}

export function AlignLeft() {
  return <Icon path="M3 6h18 M3 12h12 M3 18h15" />
}

export function AlignCenter() {
  return <Icon path="M3 6h18 M6 12h12 M4 18h16" />
}

export function AlignRight() {
  return <Icon path="M3 6h18 M9 12h12 M6 18h15" />
}

export function HorizontalRule() {
  return <Icon path="M5 12h14" />
}

export function ClearFormatting() {
  return <Icon path="M20 20H7l-5-5 9-9 9 9-5 5z M18 13l-5-5" />
}

export function Printer() {
  return <Icon path="M17 17h2a2 2 0 0 0 2-2v-4a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v4a2 2 0 0 0 2 2h2 M17 9V5a2 2 0 0 0-2-2H9a2 2 0 0 0-2 2v4 M7 15h10v6H7z" />
}

export function PageIcon() {
  return (
    <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="4" y="2" width="16" height="20" rx="2" />
      <line x1="8" y1="8" x2="16" y2="8" />
      <line x1="8" y1="12" x2="14" y2="12" />
      <line x1="8" y1="16" x2="12" y2="16" />
    </svg>
  )
}
