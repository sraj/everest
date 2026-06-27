import { useState, useRef, useEffect } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'

const SYMBOLS = {
  'Arrows':   ['→','←','↑','↓','↗','↘','⇒','⇐','⇑','⇓'],
  'Math':     ['±','×','÷','≈','≠','≤','≥','∞','√','∑','∏','∫','∂','∆','π'],
  'Currency': ['$','€','£','¥','₹','₿','¢','₽','₩','₪'],
  'Quotes':   ['"','"','\'','\'','«','»','‹','›','„','“'],
  'Dashes':   ['–','—','―','•','·','…','‽','‼','⁇','⁈'],
  'Legal':    ['©','®','™','§','¶','†','‡','№','℠','℗'],
  'Greek':    ['α','β','γ','δ','ε','θ','λ','μ','π','σ','φ','ψ','ω','Ω','Σ'],
  'Bullets':  ['●','○','■','□','◆','◇','►','▸','☞','☐','☑','☒','✗','✓','✔'],
}

interface SymbolPopoverProps {
  editor: Editor
}

export function SymbolPopover({ editor }: SymbolPopoverProps) {
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
      <ToolbarButton onClick={() => setOpen(!open)} title="Special characters">
        <span className="text-sm font-medium">Ω</span>
      </ToolbarButton>
      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 w-80 rounded-lg border border-neutral-200 bg-white shadow-lg dark:border-neutral-700 dark:bg-neutral-900 max-h-80 overflow-y-auto">
          {Object.entries(SYMBOLS).map(([category, symbols]) => (
            <div key={category} className="border-b border-neutral-100 last:border-0 dark:border-neutral-800">
              <div className="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-neutral-400 dark:text-neutral-500 bg-neutral-50 dark:bg-neutral-850">{category}</div>
              <div className="grid grid-cols-10 gap-0.5 px-1.5 pb-1.5 pt-1">
                {symbols.map((s) => (
                  <button
                    key={s}
                    onClick={() => {
                      editor.chain().focus().insertContent(s).run()
                      setOpen(false)
                    }}
                    className="w-7 h-7 flex items-center justify-center rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 text-sm transition-colors"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
