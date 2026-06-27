import { useEffect, useCallback, useState } from 'react'
import type { Editor } from '@tiptap/react'

interface ContextMenuProps {
  editor: Editor
}

interface MenuPosition {
  x: number
  y: number
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

    const clickHandler = () => setTimeout(hide, 0)
    document.addEventListener('click', clickHandler)

    const keyHandler = (e: KeyboardEvent) => { if (e.key === 'Escape') hide() }
    document.addEventListener('keydown', keyHandler)

    return () => {
      el.removeEventListener('contextmenu', handler)
      document.removeEventListener('click', clickHandler)
      document.removeEventListener('keydown', keyHandler)
    }
  }, [editor, hide])

  if (!pos) return null

  const items = [
    { label: 'Cut',       shortcut: 'Ctrl+X', action: () => { document.execCommand('cut');  hide() } },
    { label: 'Copy',      shortcut: 'Ctrl+C', action: () => { document.execCommand('copy'); hide() } },
    { label: 'Paste',     shortcut: 'Ctrl+V', action: () => { document.execCommand('paste'); hide() } },
    { label: 'Select All', shortcut: 'Ctrl+A', action: () => { editor.commands.selectAll(); hide() } },
    null,
    { label: 'Normal text', action: () => { editor.chain().focus().setParagraph().run(); hide() } },
    { label: 'Heading 1',   action: () => { editor.chain().focus().toggleHeading({ level: 1 }).run(); hide() } },
    { label: 'Heading 2',   action: () => { editor.chain().focus().toggleHeading({ level: 2 }).run(); hide() } },
    { label: 'Heading 3',   action: () => { editor.chain().focus().toggleHeading({ level: 3 }).run(); hide() } },
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

  return (
    <div
      className="fixed z-[100] min-w-[200px] rounded-lg border border-neutral-200 bg-white py-1 shadow-xl dark:border-neutral-700 dark:bg-neutral-900"
      style={{ left: pos.x, top: pos.y }}
    >
      {items.map((item, i) => {
        if (item === null) return <div key={i} className="my-1 border-t border-neutral-100 dark:border-neutral-800" />
        return (
          <button
            key={i}
            onClick={item.action}
            className="flex w-full items-center justify-between px-3 py-1.5 text-sm text-neutral-700 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800 transition-colors"
          >
            <span>{item.label}</span>
            {item.shortcut && <span className="ml-8 text-[11px] text-neutral-400 dark:text-neutral-500">{item.shortcut}</span>}
          </button>
        )
      })}
    </div>
  )
}
