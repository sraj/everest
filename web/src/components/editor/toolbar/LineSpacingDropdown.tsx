import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'
import { LINE_HEIGHTS } from '../extensions'

function LineSpacingIcon() {
  return (
    <svg className="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <path d="M3 5h18" />
      <path d="M3 12h13" />
      <path d="M3 19h18" />
      <path d="M8 9l-3 3 3 3" />
      <path d="M8 15l-3-3 3-3" />
    </svg>
  )
}

interface LineSpacingDropdownProps {
  editor: Editor
}

const LINE_HEIGHT_LABELS: Record<string, string> = {
  '1': '1',
  '1.15': '1.15',
  '1.5': '1.5',
  '2': '2',
  '2.5': '2.5',
  '3': '3',
}

export function LineSpacingDropdown({ editor }: LineSpacingDropdownProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const current = editor.getAttributes('textStyle').lineHeight || ''

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  return (
    <div ref={ref} className="relative">
      <ToolbarButton onClick={() => setOpen(!open)} active={open} title="Line spacing">
        <LineSpacingIcon />
      </ToolbarButton>

      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 min-w-[120px] rounded-lg border border-neutral-200 bg-white p-1 shadow-lg dark:border-neutral-700 dark:bg-neutral-900">
          {LINE_HEIGHTS.map((lh) => (
            <button
              key={lh}
              onClick={() => {
                editor.chain().focus().setLineHeight(lh).run()
                setOpen(false)
              }}
              className={`flex w-full items-center rounded-md px-3 py-1.5 text-sm transition-colors ${
                current === lh
                  ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                  : 'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800'
              }`}
            >
              <span>{LINE_HEIGHT_LABELS[lh]}</span>
            </button>
          ))}
          <div className="mx-2 my-1 border-t border-neutral-200 dark:border-neutral-700" />
          <button
            onClick={() => {
              editor.chain().focus().unsetLineHeight().run()
              setOpen(false)
            }}
            className="flex w-full items-center rounded-md px-3 py-1.5 text-sm text-neutral-500 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-neutral-800"
          >
            Default
          </button>
        </div>
      )}
    </div>
  )
}
