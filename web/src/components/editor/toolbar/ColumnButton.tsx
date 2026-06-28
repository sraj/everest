import { useState } from 'react'
import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'
import { Columns3 } from 'lucide-react'

interface ColumnButtonProps {
  editor: Editor
}

export function ColumnButton({ editor }: ColumnButtonProps) {
  const [cols, setCols] = useState(1)

  const cycle = () => {
    const next = cols === 1 ? 2 : cols === 2 ? 3 : 1
    setCols(next)
    const el = editor.view.dom
    if (next === 1) {
      el.removeAttribute('data-columns')
    } else {
      el.setAttribute('data-columns', String(next))
    }
  }

  return (
    <ToolbarButton onClick={cycle} active={cols > 1} title={`Columns: ${cols}`}>
      <Columns3 />
      <span className="ml-0.5 text-[10px] font-medium">{cols}</span>
    </ToolbarButton>
  )
}
