import type { Editor } from '@tiptap/react'
import { ToolbarButton } from './ToolbarButton'

interface ToCButtonProps {
  editor: Editor
}

function getHeadings(editor: Editor) {
  const headings: { level: number; text: string; pos: number }[] = []

  editor.state.doc.descendants((node, pos) => {
    if (node.type.name === 'heading') {
      headings.push({ level: node.attrs.level, text: node.textContent, pos })
    }
  })

  return headings
}

function generateToC(headings: { level: number; text: string }[]) {
  if (headings.length === 0) return '<p class="text-neutral-400 italic">No headings found in document.</p>'

  const buildTree = (items: typeof headings, startLevel: number): string => {
    let html = '<ul class="list-none pl-0 space-y-0.5">'
    let i = 0

    while (i < items.length) {
      const item = items[i]
      if (item.level < startLevel) break
      if (item.level > startLevel) {
        const subItems = items.slice(i)
        const subHtml = buildTree(subItems, startLevel + 1)
        const subCount = subItems.filter(h => h.level >= startLevel + 1).length
        html += subHtml
        i += subCount
        continue
      }
      html += `<li class="pl-${(item.level - 1) * 4}">
        <a class="text-blue-600 dark:text-blue-400 hover:underline cursor-pointer" data-toc-heading="${item.text.replace(/"/g, '&quot;')}">${item.text}</a>
      </li>`
      i++
    }

    html += '</ul>'
    return html
  }

  return buildTree(headings, 1)
}

export function ToCButton({ editor }: ToCButtonProps) {
  const handleInsertToC = () => {
    const headings = getHeadings(editor)
    const html = `<div class="table-of-contents bg-neutral-50 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 rounded-lg p-4 mb-6" data-type="toc">
      <h3 class="text-sm font-semibold uppercase tracking-wide text-neutral-500 dark:text-neutral-400 mb-3">Table of Contents</h3>
      ${generateToC(headings)}
    </div>`

    editor.chain().focus().insertContent(html).run()
  }

  return (
    <ToolbarButton onClick={handleInsertToC} title="Insert table of contents">
      <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="8" y1="6" x2="21" y2="6" />
        <line x1="8" y1="12" x2="21" y2="12" />
        <line x1="8" y1="18" x2="21" y2="18" />
        <line x1="3" y1="6" x2="3.01" y2="6" />
        <line x1="3" y1="12" x2="3.01" y2="12" />
        <line x1="3" y1="18" x2="3.01" y2="18" />
      </svg>
    </ToolbarButton>
  )
}
