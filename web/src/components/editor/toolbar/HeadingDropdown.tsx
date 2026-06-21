import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { Paragraph, Heading1, Heading2, Heading3, Heading4, Heading5, Heading6, ChevronDown } from '../icons'

interface HeadingOption {
  label: string
  level: number | null
  icon: React.ComponentType<{ className?: string }>
}

const HEADING_OPTIONS: HeadingOption[] = [
  { label: 'Paragraph', level: null, icon: Paragraph },
  { label: 'Heading 1', level: 1, icon: Heading1 },
  { label: 'Heading 2', level: 2, icon: Heading2 },
  { label: 'Heading 3', level: 3, icon: Heading3 },
  { label: 'Heading 4', level: 4, icon: Heading4 },
  { label: 'Heading 5', level: 5, icon: Heading5 },
  { label: 'Heading 6', level: 6, icon: Heading6 },
]

export function HeadingDropdown({ editor }: { editor: Editor }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [open])

  const activeOption = HEADING_OPTIONS.find(o =>
    o.level !== null
      ? editor.isActive('heading', { level: o.level })
      : !editor.isActive('heading')
  )

  const label = activeOption?.label ?? 'Paragraph'

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="inline-flex items-center justify-between gap-1 h-8 min-w-[6.5rem] px-2 rounded-md text-sm font-normal transition-colors
          hover:bg-neutral-100 dark:hover:bg-neutral-800
          text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white"
      >
        <span className="truncate">{label}</span>
        <ChevronDown className="size-3 shrink-0 text-neutral-400" />
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 w-40 rounded-lg border border-neutral-200 bg-white py-1 shadow-lg dark:border-neutral-700 dark:bg-neutral-900">
          {HEADING_OPTIONS.map((opt) => {
            const isActive = opt.level !== null
              ? editor.isActive('heading', { level: opt.level })
              : !editor.isActive('heading')

            return (
              <button
                key={opt.label}
                onClick={() => {
                  if (opt.level) {
                    editor.chain().focus().toggleHeading({ level: opt.level }).run()
                  } else {
                    editor.chain().focus().setParagraph().run()
                  }
                  setOpen(false)
                }}
                className={`flex items-center gap-2 w-full px-3 py-1.5 text-sm transition-colors ${
                  isActive
                    ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400'
                    : 'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800'
                }`}
              >
                <opt.icon className="size-4 shrink-0" />
                <span>{opt.label}</span>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
