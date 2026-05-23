import { useState, useRef, useEffect, useCallback } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'
import { Color as ColorIcon } from '../icons'

interface ColorPopoverProps {
  editor: Editor
}

const TEXT_COLORS = [
  '#000000', '#434343', '#666666', '#999999', '#b7b7b7', '#cccccc',
  '#ff0000', '#e74c3c', '#e67e22', '#f39c12', '#2ecc71', '#27ae60',
  '#3498db', '#2980b9', '#9b59b6', '#8e44ad',
]

const HIGHLIGHT_COLORS = [
  '#ffff00', '#ffd700', '#ffa500', '#ff6b6b', '#98fb98', '#87ceeb',
  '#dda0dd', '#ffb6c1', '#e0e0e0', '#ffffff',
]

const POPOVER_WIDTH = 216

function ColorGrid({ colors, selected, onSelect }: {
  colors: string[]
  selected: string | null
  onSelect: (color: string) => void
}) {
  return (
    <div className="grid grid-cols-8 gap-1">
      {colors.map((color) => (
        <button
          key={color}
          onClick={() => onSelect(color)}
          className={`size-6 rounded border-2 ${
            selected === color ? 'border-blue-500' : 'border-transparent'
          }`}
          style={{ backgroundColor: color }}
        />
      ))}
      <button
        onClick={() => onSelect('')}
        className="size-6 rounded border border-neutral-300 text-[10px] text-neutral-400 hover:bg-neutral-100 dark:border-neutral-600 dark:hover:bg-neutral-800"
        title="Remove color"
      >
        ∅
      </button>
    </div>
  )
}

export function ColorPopover({ editor }: ColorPopoverProps) {
  const [open, setOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const popoverRef = useRef<HTMLDivElement>(null)
  const [style, setStyle] = useState<React.CSSProperties>({})

  const recalcPosition = useCallback(() => {
    if (!buttonRef.current) return
    const rect = buttonRef.current.getBoundingClientRect()
    const left = Math.min(rect.left, window.innerWidth - POPOVER_WIDTH - 8)
    setStyle({
      position: 'fixed',
      top: rect.bottom + 4,
      left: Math.max(8, left),
      zIndex: 50,
      width: POPOVER_WIDTH,
    })
  }, [])

  useEffect(() => {
    if (open) {
      recalcPosition()
      const onResize = () => recalcPosition()
      window.addEventListener('scroll', onResize, true)
      window.addEventListener('resize', onResize)
      return () => {
        window.removeEventListener('scroll', onResize, true)
        window.removeEventListener('resize', onResize)
      }
    }
  }, [open, recalcPosition])

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        popoverRef.current &&
        buttonRef.current &&
        !popoverRef.current.contains(e.target as Node) &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  const currentColor = editor.getAttributes('textStyle').color || null
  const currentHighlight = editor.getAttributes('highlight').color || null

  return (
    <>
      <ToolbarButton
        ref={buttonRef}
        onClick={() => setOpen(!open)}
        active={!!currentColor || !!currentHighlight}
        title="Text color / Highlight"
      >
        <ColorIcon />
      </ToolbarButton>

      {open && (
        <div ref={popoverRef} style={style} className="rounded-lg border border-neutral-200 bg-white p-2.5 shadow-lg dark:border-neutral-700 dark:bg-neutral-900">
          <p className="mb-2 text-[10px] font-medium uppercase tracking-wider text-neutral-500">Color / Highlight</p>
          <div className="space-y-2.5">
            <div>
              <p className="mb-1 text-[10px] text-neutral-400">Text</p>
              <ColorGrid colors={TEXT_COLORS} selected={currentColor} onSelect={(color) => {
                if (color) editor.chain().focus().setColor(color).run()
                else editor.chain().focus().unsetColor().run()
              }} />
            </div>
            <div>
              <p className="mb-1 text-[10px] text-neutral-400">Highlight</p>
              <ColorGrid colors={HIGHLIGHT_COLORS} selected={currentHighlight} onSelect={(color) => {
                if (color) editor.chain().focus().toggleHighlight({ color }).run()
                else editor.chain().focus().toggleHighlight().run()
              }} />
            </div>
          </div>
        </div>
      )}
    </>
  )
}
