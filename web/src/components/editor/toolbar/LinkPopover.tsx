import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'
import { Link as LinkIcon } from '../icons'

interface LinkPopoverProps {
  editor: Editor
}

export function LinkPopover({ editor }: LinkPopoverProps) {
  const [open, setOpen] = useState(false)
  const [url, setUrl] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const popoverRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open) {
      const previousUrl = editor.getAttributes('link').href || ''
      setUrl(previousUrl)
      setTimeout(() => inputRef.current?.select(), 0)
    }
  }, [open, editor])

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  const handleSubmit = () => {
    if (url) {
      editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
    } else {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
    }
    setOpen(false)
  }

  const handleRemove = () => {
    editor.chain().focus().extendMarkRange('link').unsetLink().run()
    setOpen(false)
  }

  return (
    <div ref={popoverRef} className="relative">
      <ToolbarButton
        onClick={() => setOpen(!open)}
        active={editor.isActive('link')}
        title="Link"
      >
        <LinkIcon />
      </ToolbarButton>

      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 flex items-center gap-1.5 rounded-lg border border-neutral-200 bg-white p-1.5 shadow-lg dark:border-neutral-700 dark:bg-neutral-900">
          <input
            ref={inputRef}
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
            placeholder="https://..."
            className="w-48 rounded-md border border-neutral-300 bg-transparent px-2 py-1 text-sm text-neutral-900 placeholder-neutral-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-neutral-600 dark:text-neutral-100"
          />
          <button
            onClick={handleSubmit}
            className="rounded-md bg-blue-600 px-2 py-1 text-xs text-white hover:bg-blue-700"
          >
            Apply
          </button>
          <button
            onClick={handleRemove}
            className="rounded-md px-2 py-1 text-xs text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800"
          >
            Remove
          </button>
        </div>
      )}
    </div>
  )
}
