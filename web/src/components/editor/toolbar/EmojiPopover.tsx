import { useState, useRef, useEffect } from 'react'
import { ToolbarButton } from './ToolbarButton'
import { ChevronDown } from '../icons'

const EMOJIS = ['😀','😂','❤️','👍','🔥','🎉','✨','🚀','💡','📄','✏️','🔍','📎','📌','⭐','💪','🤔','👀','✅','❌','⚡','🎯','💻','📱','🌍']

interface EmojiPopoverProps {
  editor: any
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
        <div className="absolute bottom-full left-0 mb-1 z-50 rounded-lg border border-neutral-200 bg-white p-2 shadow-lg dark:border-neutral-700 dark:bg-neutral-900 grid grid-cols-5 gap-1">
          {EMOJIS.map((emoji) => (
            <button
              key={emoji}
              onClick={() => {
                editor.chain().focus().insertContent(emoji).run()
                setOpen(false)
              }}
              className="w-8 h-8 flex items-center justify-center rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 text-lg transition-colors"
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
