import type { Editor } from '@tiptap/react'
import { useRef, useCallback } from 'react'
import { ToolbarButton, ToolbarDivider } from './ToolbarButton'
import { LinkPopover } from './LinkPopover'
import { ColorPopover } from './ColorPopover'
import { PageSizePopover } from './PageSizePopover'
import { FontFamilyDropdown } from './FontFamilyDropdown'
import { FontSizeDropdown } from './FontSizeDropdown'
import { LineSpacingDropdown } from './LineSpacingDropdown'
import { HeadingDropdown } from './HeadingDropdown'
import { ALL_PAGE_SIZES } from '../extensions'
import type { PageSizeKey } from '../extensions'
import {
  Undo, Redo, Bold, Italic, Underline, Strikethrough,
  SuperscriptIcon, SubscriptIcon, Image as ImageIcon,
  BulletList, OrderedList, TaskList,
  Blockquote, Code,
  AlignLeft, AlignCenter, AlignRight, Printer,
  HorizontalRule, ClearFormatting,
} from '../icons'

interface ToolbarProps {
  editor: Editor
  pageSize: PageSizeKey
  onPageSizeChange: (size: PageSizeKey) => void
}

function printStyles(pageSize: PageSizeKey) {
  const s = ALL_PAGE_SIZES[pageSize]
  const w = Math.round(s.pageWidth * 25.4 / 96)
  const h = Math.round(s.pageHeight * 25.4 / 96)
  return `
    @page { size: ${w}mm ${h}mm; margin: 0; }
    @media print {
      body { print-color-adjust: exact; -webkit-print-color-adjust: exact; }
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      color: #000; line-height: 1.6; padding: 40px;
    }
    h1 { font-size: 28px; font-weight: 700; margin: 32px 0 16px; line-height: 1.2; }
    h2 { font-size: 24px; font-weight: 600; margin: 24px 0 12px; line-height: 1.3; }
    h3 { font-size: 20px; font-weight: 600; margin: 20px 0 8px; line-height: 1.4; }
    h4 { font-size: 18px; font-weight: 500; margin: 16px 0 8px; }
    p { margin-bottom: 16px; }
    ul, ol { padding-left: 24px; margin: 16px 0; }
    li { margin-bottom: 4px; }
    ul[data-type="taskList"] { list-style: none; padding-left: 0; }
    ul[data-type="taskList"] li { display: flex; align-items: flex-start; gap: 8px; margin: 4px 0; }
    blockquote {
      border-left: 4px solid #ccc; padding-left: 16px; margin: 16px 0;
      font-style: italic; color: #555;
    }
    pre {
      background: #1e1e1e; color: #e0e0e0; padding: 16px; border-radius: 8px;
      margin: 16px 0; overflow-x: auto; font-family: monospace; font-size: 14px;
    }
    code { background: #f0f0f0; color: #d63384; padding: 2px 6px; border-radius: 4px; font-size: 14px; font-family: monospace; }
    pre code { background: transparent; color: inherit; padding: 0; }
    img { max-width: 100%; height: auto; margin: 16px 0; }
    a { color: #2563eb; text-decoration: underline; }
    hr { border: none; border-top: 1px solid #ddd; margin: 32px 0; }
    mark { background: #ffd700; }
    table { width: 100%; border-collapse: collapse; margin: 16px 0; }
    th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
    th { background: #f5f5f5; }
  `
}

export function Toolbar({ editor, pageSize, onPageSizeChange }: ToolbarProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handlePrint = useCallback(() => {
    const content = editor.getHTML()
    const win = window.open('', '_blank', 'width=800,height=600')
    if (!win) return

    win.document.write(`<!DOCTYPE html><html><head><title>Print</title><style>${printStyles(pageSize)}</style></head><body>${content}<script>window.onload=function(){window.onafterprint=function(){window.close()};window.print()}<\/script></body></html>`)
    win.document.close()
  }, [editor, pageSize])

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (event) => {
      const url = event.target?.result as string
      editor.chain().focus().setImage({ src: url }).run()
    }
    reader.readAsDataURL(file)
    e.target.value = ''
  }

  return (
    <>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleImageUpload}
      />

      <div className="flex flex-wrap items-center gap-0.5 border-b border-neutral-200 bg-neutral-50 px-2 py-1.5 dark:border-neutral-800 dark:bg-neutral-900">
        {/* Undo / Redo */}
        <ToolbarButton onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().undo()} title="Undo">
          <Undo />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().redo()} title="Redo">
          <Redo />
        </ToolbarButton>

        <ToolbarDivider />

        {/* Font family / size / line spacing */}
        <FontFamilyDropdown editor={editor} />
        <FontSizeDropdown editor={editor} />
        <LineSpacingDropdown editor={editor} />

        <ToolbarDivider />

        {/* Text formatting */}
        <ToolbarButton onClick={() => editor.chain().focus().toggleBold().run()} active={editor.isActive('bold')} title="Bold">
          <Bold />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleItalic().run()} active={editor.isActive('italic')} title="Italic">
          <Italic />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleUnderline().run()} active={editor.isActive('underline')} title="Underline">
          <Underline />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleStrike().run()} active={editor.isActive('strike')} title="Strikethrough">
          <Strikethrough />
        </ToolbarButton>

        <ToolbarButton onClick={() => editor.chain().focus().toggleSuperscript().run()} active={editor.isActive('superscript')} title="Superscript">
          <SuperscriptIcon />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleSubscript().run()} active={editor.isActive('subscript')} title="Subscript">
          <SubscriptIcon />
        </ToolbarButton>

        <ToolbarDivider />

        {/* Color / Highlight */}
        <ColorPopover editor={editor} />

        {/* Link */}
        <LinkPopover editor={editor} />

        {/* Image */}
        <ToolbarButton onClick={() => fileInputRef.current?.click()} title="Image">
          <ImageIcon />
        </ToolbarButton>

        <ToolbarDivider />

        {/* Headings */}
        <HeadingDropdown editor={editor} />

        <ToolbarDivider />

        {/* Lists */}
        <ToolbarButton onClick={() => editor.chain().focus().toggleBulletList().run()} active={editor.isActive('bulletList')} title="Bullet List">
          <BulletList />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleOrderedList().run()} active={editor.isActive('orderedList')} title="Numbered List">
          <OrderedList />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleTaskList().run()} active={editor.isActive('taskList')} title="Task List">
          <TaskList />
        </ToolbarButton>

        <ToolbarButton onClick={() => editor.chain().focus().toggleBlockquote().run()} active={editor.isActive('blockquote')} title="Quote">
          <Blockquote />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleCodeBlock().run()} active={editor.isActive('codeBlock')} title="Code Block">
          <Code />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setHorizontalRule().run()} title="Horizontal Rule">
          <HorizontalRule />
        </ToolbarButton>

        <ToolbarDivider />

        {/* Alignment */}
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign('left').run()} active={editor.isActive({ textAlign: 'left' })} title="Align Left">
          <AlignLeft />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign('center').run()} active={editor.isActive({ textAlign: 'center' })} title="Align Center">
          <AlignCenter />
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign('right').run()} active={editor.isActive({ textAlign: 'right' })} title="Align Right">
          <AlignRight />
        </ToolbarButton>

        <ToolbarDivider />

        {/* Clear formatting */}
        <ToolbarButton onClick={() => editor.chain().focus().clearNodes().unsetAllMarks().run()} title="Clear formatting">
          <ClearFormatting />
        </ToolbarButton>

        {/* Print */}
        <ToolbarButton onClick={handlePrint} title="Print">
          <Printer />
        </ToolbarButton>

        {/* Page size */}
        <PageSizePopover pageSize={pageSize} onPageSizeChange={onPageSizeChange} />
      </div>
    </>
  )
}
