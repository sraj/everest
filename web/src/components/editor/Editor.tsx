import { useEditor, EditorContent } from '@tiptap/react'
import { useRef, useEffect, useState } from 'react'
import { createExtensions } from './extensions'
import type { PageSizeKey } from './extensions'
import { Toolbar } from './toolbar/Toolbar'
import { ContextMenu } from './ContextMenu'
import './editor.css'

interface EditorProps {
  content?: string
  onChange?: (html: string) => void
  editable?: boolean
}

function EditorInner({ content = '', onChange, editable = true, pageSize, onPageSizeChange }: EditorProps & {
  pageSize: PageSizeKey
  onPageSizeChange: (size: PageSizeKey) => void
}) {
  const editor = useEditor({
    extensions: createExtensions(pageSize),
    content: content || '',
    editable,
    immediatelyRender: false,
    shouldRerenderOnTransaction: false,
    autofocus: 'end',
    onUpdate: ({ editor }) => {
      onChange?.(editor.getHTML())
    },
    editorProps: {
      attributes: {
        class: 'tiptap-editor prose prose-sm sm:prose lg:prose-lg xl:prose-2xl mx-auto focus:outline-none min-h-[200px]',
      },
    },
  })

  if (!editor) return null

  return (
    <div className="flex flex-col h-full rounded-xl border border-neutral-200 bg-white shadow-sm dark:border-neutral-800 dark:bg-neutral-950">
      <Toolbar editor={editor} pageSize={pageSize} onPageSizeChange={onPageSizeChange} />
      <div className="flex-1 overflow-hidden">
        <div className="h-full overflow-y-auto bg-neutral-100 dark:bg-neutral-900">
          <EditorContent editor={editor} className="h-full" />
        </div>
      </div>
      <div className="border-t border-neutral-200 dark:border-neutral-700 px-4 py-1.5 flex items-center gap-4 text-[11px] text-neutral-400 dark:text-neutral-500">
        <span>{editor.storage.characterCount?.characters?.() ?? 0} characters</span>
        <span>{editor.storage.characterCount?.words?.() ?? 0} words</span>
      </div>
      <ContextMenu editor={editor} />
    </div>
  )
}

export function Editor({ content = '', onChange, editable = true }: EditorProps) {
  const [pageSize, setPageSize] = useState<PageSizeKey>('A4')
  const contentRef = useRef(content)

  useEffect(() => {
    contentRef.current = content
  }, [content])

  return (
    <EditorInner
      key={pageSize}
      content={contentRef.current}
      onChange={onChange}
      editable={editable}
      pageSize={pageSize}
      onPageSizeChange={setPageSize}
    />
  )
}
