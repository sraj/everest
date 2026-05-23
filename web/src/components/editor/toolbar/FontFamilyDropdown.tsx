import type { Editor } from '@tiptap/react'
import { FONT_FAMILIES } from '../extensions'

interface FontFamilyDropdownProps {
  editor: Editor
}

export function FontFamilyDropdown({ editor }: FontFamilyDropdownProps) {
  const current = editor.getAttributes('textStyle').fontFamily || ''

  return (
    <select
      value={current}
      onChange={(e) => {
        const val = e.target.value
        if (val) editor.chain().focus().setFontFamily(val).run()
        else editor.chain().focus().unsetFontFamily().run()
      }}
      className="h-8 rounded-md border border-neutral-200 bg-transparent px-1.5 text-xs text-neutral-700 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-neutral-700 dark:text-neutral-300 max-w-[120px]"
    >
      <option value="">Normal</option>
      {FONT_FAMILIES.map((f) => (
        <option key={f} value={f} style={{ fontFamily: f }}>
          {f}
        </option>
      ))}
    </select>
  )
}
