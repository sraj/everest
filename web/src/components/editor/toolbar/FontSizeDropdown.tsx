import type { Editor } from '@tiptap/react'
import { FONT_SIZES } from '../extensions'

interface FontSizeDropdownProps {
  editor: Editor
}

export function FontSizeDropdown({ editor }: FontSizeDropdownProps) {
  const current = editor.getAttributes('textStyle').fontSize || ''

  return (
    <select
      value={current}
      onChange={(e) => {
        const val = e.target.value
        if (val) editor.chain().focus().setFontSize(val).run()
        else editor.chain().focus().unsetFontSize().run()
      }}
      className="h-8 rounded-md border border-neutral-200 bg-transparent px-1.5 text-xs text-neutral-700 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-neutral-700 dark:text-neutral-300 w-16"
    >
      <option value="">Size</option>
      {FONT_SIZES.map((s) => (
        <option key={s} value={`${s}px`}>
          {s}
        </option>
      ))}
    </select>
  )
}
