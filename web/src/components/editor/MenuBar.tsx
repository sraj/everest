import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { PAGE_SIZE_LABELS } from './extensions'
import type { PageSizeKey } from './extensions'

interface MenuBarProps {
  editor: Editor
  onOpenFind: () => void
  onInsertTable: () => void
  onInsertToC: () => void
  onInsertEmoji: () => void
  onInsertSymbol: () => void
  pageSize: PageSizeKey
}

interface MenuProps {
  label: string
  children: React.ReactNode
}

function Menu({ label, children }: MenuProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        onMouseEnter={() => { if (document.querySelector('[data-menu-open]')) setOpen(true) }}
        className={`px-2 py-0.5 text-[13px] rounded transition-colors ${open ? 'bg-neutral-200 dark:bg-neutral-700 text-neutral-900 dark:text-white' : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'}`}
      >
        {label}
      </button>
      {open && (
        <div
          data-menu-open
          className="absolute top-full left-0 mt-0.5 z-50 min-w-[200px] rounded-lg border border-neutral-200 bg-white py-1 shadow-xl dark:border-neutral-700 dark:bg-neutral-900"
          onMouseLeave={() => setOpen(false)}
        >
          {children}
        </div>
      )}
    </div>
  )
}

function MenuItem({ label, shortcut, onClick, disabled, danger }: { label: string; shortcut?: string; onClick: () => void; disabled?: boolean; danger?: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`flex w-full items-center justify-between px-3 py-1 text-[13px] transition-colors ${
        disabled ? 'opacity-30 cursor-default' :
        danger ? 'text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20' :
        'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800'
      }`}
    >
      <span>{label}</span>
      {shortcut && <span className="ml-8 text-[11px] text-neutral-400 dark:text-neutral-500">{shortcut}</span>}
    </button>
  )
}

function MenuDivider() {
  return <div className="my-1 border-t border-neutral-100 dark:border-neutral-800" />
}

export function MenuBar({ editor, onOpenFind, onInsertTable, onInsertToC: _onInsertToC, onInsertEmoji: _e, onInsertSymbol: _s, pageSize }: MenuBarProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const print = () => {
    const content = editor.getHTML()
    const win = window.open('', '_blank', 'width=800,height=600')
    if (!win) return
    win.document.write(`<!DOCTYPE html><html><head><title>Print</title></head><body>${content}<script>window.onload=function(){window.onafterprint=function(){window.close()};window.print()}<\/script></body></html>`)
    win.document.close()
  }

  const downloadHTML = () => {
    const content = editor.getHTML()
    const blob = new Blob([`<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>${content}</body></html>`], { type: 'text/html' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'document.html'
    a.click()
    URL.revokeObjectURL(url)
  }

  const imageUpload = () => {
    const el = fileInputRef.current
    if (!el) return
    const file = el.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (e) => {
      const url = e.target?.result as string
      editor.chain().focus().setImage({ src: url }).run()
    }
    reader.readAsDataURL(file)
    el.value = ''
  }

  return (
    <>
      <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={imageUpload} />
      <div className="flex items-center gap-0.5 bg-white px-2 py-0.5 border-b border-neutral-200 dark:bg-neutral-950 dark:border-neutral-800">
        <Menu label="File">
          <MenuItem label="Print" shortcut="⌘P" onClick={print} />
          <MenuItem label="Download as HTML" onClick={downloadHTML} />
          <MenuDivider />
          <MenuItem label="Page setup" onClick={() => {}} disabled />
        </Menu>

        <Menu label="Edit">
          <MenuItem label="Undo" shortcut="⌘Z" onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().undo()} />
          <MenuItem label="Redo" shortcut="⌘Y" onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().redo()} />
          <MenuDivider />
          <MenuItem label="Cut" shortcut="⌘X" onClick={() => document.execCommand('cut')} disabled={editor.state.selection.empty} />
          <MenuItem label="Copy" shortcut="⌘C" onClick={() => document.execCommand('copy')} disabled={editor.state.selection.empty} />
          <MenuItem label="Paste" shortcut="⌘V" onClick={() => document.execCommand('paste')} />
          <MenuItem label="Select all" shortcut="⌘A" onClick={() => editor.commands.selectAll()} />
          <MenuDivider />
          <MenuItem label="Find and replace" shortcut="⌘F" onClick={onOpenFind} />
        </Menu>

        <Menu label="View">
          <MenuItem label={`Page size: ${PAGE_SIZE_LABELS[pageSize]}`} onClick={() => {}} disabled />
          <MenuItem label="Zoom" onClick={() => {}} disabled />
        </Menu>

        <Menu label="Insert">
          <MenuItem label="Image" onClick={() => fileInputRef.current?.click()} />
          <MenuItem label="Table" onClick={onInsertTable} />
          <MenuItem label="Link" shortcut="⌘K" onClick={() => {
            const url = window.prompt('Enter URL:')
            if (url) editor.chain().focus().setLink({ href: url }).run()
          }} />
          <MenuDivider />
          <MenuItem label="Horizontal rule" onClick={() => editor.chain().focus().setHorizontalRule().run()} />
          <MenuItem label="Emoji" onClick={() => {}} disabled />
          <MenuItem label="Special characters" onClick={() => {}} disabled />
          <MenuDivider />
          <MenuItem label="Table of contents" onClick={() => {
            const headings: { level: number; text: string }[] = []
            editor.state.doc.descendants((node) => {
              if (node.type.name === 'heading') headings.push({ level: node.attrs.level, text: node.textContent })
            })
            if (headings.length === 0) return
            const html = `<div class="table-of-contents bg-neutral-50 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 rounded-lg p-4 mb-6" data-type="toc"><h3 class="text-sm font-semibold uppercase tracking-wide text-neutral-500 dark:text-neutral-400 mb-3">Table of Contents</h3><ul class="list-none pl-0 space-y-0.5">${headings.map(h => `<li class="pl-${(h.level - 1) * 4}"><a class="text-blue-600 dark:text-blue-400 hover:underline cursor-pointer" data-toc-heading="${h.text.replace(/"/g, '&quot;')}">${h.text}</a></li>`).join('')}</ul></div>`
            editor.chain().focus().insertContent(html).run()
          }} />
        </Menu>

        <Menu label="Format">
          <MenuItem label="Bold" shortcut="⌘B" onClick={() => editor.chain().focus().toggleBold().run()} />
          <MenuItem label="Italic" shortcut="⌘I" onClick={() => editor.chain().focus().toggleItalic().run()} />
          <MenuItem label="Underline" shortcut="⌘U" onClick={() => editor.chain().focus().toggleUnderline().run()} />
          <MenuItem label="Strikethrough" onClick={() => editor.chain().focus().toggleStrike().run()} />
          <MenuDivider />
          <MenuItem label="Superscript" onClick={() => editor.chain().focus().toggleSuperscript().run()} />
          <MenuItem label="Subscript" onClick={() => editor.chain().focus().toggleSubscript().run()} />
          <MenuDivider />
          <MenuItem label="Align left" onClick={() => editor.chain().focus().setTextAlign('left').run()} />
          <MenuItem label="Align center" onClick={() => editor.chain().focus().setTextAlign('center').run()} />
          <MenuItem label="Align right" onClick={() => editor.chain().focus().setTextAlign('right').run()} />
          <MenuDivider />
          <MenuItem label="Clear formatting" onClick={() => editor.chain().focus().clearNodes().unsetAllMarks().run()} />
        </Menu>
      </div>
    </>
  )
}
