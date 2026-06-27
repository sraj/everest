import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'

const EMOJIS = [
  '😀','😂','🤣','😍','🥰','😎','🤩','😇','🤔','🙄',
  '😢','😡','🥺','😴','🤯','🥳','😱','🤗','🫡','🫠',
  '👍','👎','👏','🙌','💪','🤝','✌️','🤞','🫶','👋',
  '❤️','🧡','💛','💚','💙','💜','🖤','🤍','🤎','💔',
  '🔥','⭐','✨','💡','🚀','🎉','🎯','✅','❌','⚡',
  '📄','📎','📌','✏️','🔍','📊','📈','💻','📱','🌍',
  '0️⃣','1️⃣','2️⃣','3️⃣','4️⃣','5️⃣','6️⃣','7️⃣','8️⃣','9️⃣',
  '→','←','↑','↓','↗️','↘️','◀️','▶️','🔄','🔗',
]

interface EmojiPopoverProps {
  editor: Editor
}

export function EmojiPopover({ editor }: EmojiPopoverProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    if (open) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  return (
    <div ref={ref} className="relative">
      <ToolbarButton onClick={() => setOpen(!open)} title="Emoji">
        <span className="text-sm">😊</span>
      </ToolbarButton>
      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 w-72 rounded-lg border border-neutral-200 bg-white p-2 shadow-lg dark:border-neutral-700 dark:bg-neutral-900 grid grid-cols-10 gap-0.5 max-h-64 overflow-y-auto">
          {EMOJIS.map((emoji) => (
            <button
              key={emoji}
              onClick={() => {
                editor.chain().focus().insertContent(emoji).run()
                setOpen(false)
              }}
              className="w-7 h-7 flex items-center justify-center rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 text-base transition-colors"
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
