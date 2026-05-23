import { useState, useRef, useEffect } from 'react'
import { ToolbarButton } from './ToolbarButton'
import { PageIcon } from '../icons'
import { ALL_PAGE_SIZES, PAGE_SIZE_LABELS, PAGE_SIZE_ORDER } from '../extensions'
import type { PageSizeKey } from '../extensions'

interface PageSizePopoverProps {
  pageSize: PageSizeKey
  onPageSizeChange: (size: PageSizeKey) => void
}

function formatDimensions(size: PageSizeKey): string {
  const s = ALL_PAGE_SIZES[size]
  return `${s.pageWidth}×${s.pageHeight}`
}

export function PageSizePopover({ pageSize, onPageSizeChange }: PageSizePopoverProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

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
      <ToolbarButton
        onClick={() => setOpen(!open)}
        active={open}
        title={`Page size: ${PAGE_SIZE_LABELS[pageSize]}`}
      >
        <PageIcon />
        <span className="ml-1 text-[10px] font-medium">{PAGE_SIZE_LABELS[pageSize]}</span>
      </ToolbarButton>

      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 rounded-lg border border-neutral-200 bg-white p-1.5 shadow-lg dark:border-neutral-700 dark:bg-neutral-900 min-w-[160px]">
          {PAGE_SIZE_ORDER.map((size) => (
            <button
              key={size}
              onClick={() => {
                onPageSizeChange(size)
                setOpen(false)
              }}
              className={`flex w-full items-center justify-between rounded-md px-3 py-1.5 text-sm transition-colors ${
                size === pageSize
                  ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                  : 'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800'
              }`}
            >
              <span className="font-medium">{PAGE_SIZE_LABELS[size]}</span>
              <span className="text-[10px] text-neutral-400">{formatDimensions(size)}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
