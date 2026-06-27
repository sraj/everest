import { useEffect, useCallback, useState } from 'react'
import type { Editor } from '@tiptap/react'

interface ContextMenuProps {
  editor: Editor
}

interface MenuPosition {
  x: number
  y: number
}

interface MenuItem {
  label: string
  shortcut?: string
  action: () => void
  danger?: boolean
}

export function ContextMenu({ editor }: ContextMenuProps) {
  const [pos, setPos] = useState<MenuPosition | null>(null)

  const hide = useCallback(() => setPos(null), [])

  useEffect(() => {
    const el = editor.view.dom
    const handler = (e: MouseEvent) => {
      e.preventDefault()
      setPos({ x: e.clientX, y: e.clientY })
    }
    el.addEventListener('contextmenu', handler)
    document.addEventListener('click', () => setTimeout(hide, 0))

    const keyHandler = (e: KeyboardEvent) => { if (e.key === 'Escape') hide() }
    document.addEventListener('keydown', keyHandler)

    return () => {
      el.removeEventListener('contextmenu', handler)
      document.removeEventListener('keydown', keyHandler)
    }
  }, [editor, hide])

  if (!pos) return null

  const isTable = editor.isActive('table')
  const isLink = editor.isActive('link')
  const isImage = editor.isActive('image')
  const isCodeBlock = editor.isActive('codeBlock')
  const isBlockquote = editor.isActive('blockquote')

  const hasSelection = !editor.state.selection.empty

  let items: (MenuItem | null)[] = [
    { label: 'Cut',   shortcut: '⌘X', action: () => { document.execCommand('cut');  hide() } },
    { label: 'Copy',  shortcut: '⌘C', action: () => { document.execCommand('copy'); hide() } },
    { label: 'Paste', shortcut: '⌘V', action: () => { document.execCommand('paste'); hide() } },
    hasSelection ? null : { label: 'Select All', shortcut: '⌘A', action: () => { editor.commands.selectAll(); hide() } },
  ]

  if (isTable) {
    items = [
      { label: 'Insert row above',    action: () => { editor.chain().focus().addRowBefore().run(); hide() } },
      { label: 'Insert row below',    action: () => { editor.chain().focus().addRowAfter().run(); hide() } },
      { label: 'Insert column before', action: () => { editor.chain().focus().addColumnBefore().run(); hide() } },
      { label: 'Insert column after',  action: () => { editor.chain().focus().addColumnAfter().run(); hide() } },
      null,
      { label: 'Delete row',    action: () => { editor.chain().focus().deleteRow().run(); hide() }, danger: true },
      { label: 'Delete column', action: () => { editor.chain().focus().deleteColumn().run(); hide() }, danger: true },
      { label: 'Delete table',  action: () => { editor.chain().focus().deleteTable().run(); hide() }, danger: true },
    ]
  } else if (isLink) {
    items = [
      { label: 'Open link',   action: () => { window.open(editor.getAttributes('link').href, '_blank'); hide() } },
      { label: 'Edit link',   action: hide },
      null,
      { label: 'Remove link', action: () => { editor.chain().focus().unsetLink().run(); hide() }, danger: true },
    ]
  } else if (isImage) {
    items = [
      ...items,
      null,
      { label: 'Remove image', action: () => { editor.chain().focus().deleteSelection().run(); hide() }, danger: true },
    ]
  } else {
    items = [
      ...items,
      null,
      { label: 'Normal text',   action: () => { editor.chain().focus().setParagraph().run(); hide() } },
      { label: 'Heading 1',     action: () => { editor.chain().focus().toggleHeading({ level: 1 }).run(); hide() } },
      { label: 'Heading 2',     action: () => { editor.chain().focus().toggleHeading({ level: 2 }).run(); hide() } },
      { label: 'Heading 3',     action: () => { editor.chain().focus().toggleHeading({ level: 3 }).run(); hide() } },
      null,
      { label: 'Bullet list',    action: () => { editor.chain().focus().toggleBulletList().run(); hide() } },
      { label: 'Numbered list',  action: () => { editor.chain().focus().toggleOrderedList().run(); hide() } },
      { label: 'Task list',      action: () => { editor.chain().focus().toggleTaskList().run(); hide() } },
      null,
      { label: 'Blockquote', action: () => { editor.chain().focus().toggleBlockquote().run(); hide() } },
      { label: 'Code block',  action: () => { editor.chain().focus().toggleCodeBlock().run(); hide() } },
      null,
      { label: 'Clear formatting', action: () => { editor.chain().focus().clearNodes().unsetAllMarks().run(); hide() } },
    ]
  }

  const menuWidth = isTable ? 220 : 200

  return (
    <>
      <div className="fixed inset-0 z-[99]" onClick={hide} />
      <div
        className="fixed z-[100] rounded-lg border border-neutral-200 bg-white py-1 shadow-lg shadow-neutral-200/50 dark:border-neutral-700 dark:bg-neutral-900 dark:shadow-neutral-950"
        style={{ left: Math.min(pos.x, window.innerWidth - menuWidth - 8), top: Math.min(pos.y, window.innerHeight - items.filter(Boolean).length * 35 - 16), minWidth: menuWidth }}
      >
        {items.map((item, i) => {
          if (item === null) return <div key={i} className="my-1 border-t border-neutral-100 dark:border-neutral-800" />
          return (
            <button
              key={i}
              onClick={item.action}
              className={`flex w-full items-center justify-between px-3 py-1.5 text-[13px] transition-colors ${
                item.danger
                  ? 'text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20'
                  : 'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800'
              }`}
            >
              <span className="font-medium">{item.label}</span>
              {item.shortcut && <span className="ml-8 text-[11px] text-neutral-400 dark:text-neutral-500">{item.shortcut}</span>}
            </button>
          )
        })}
      </div>
    </>
  )
}
