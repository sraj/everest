import { useEditor, EditorContent } from '@tiptap/react'
import { useRef, useEffect, useState } from 'react'
import { createExtensions } from './extensions'
import type { PageSizeKey } from './extensions'
import { Toolbar } from './toolbar/Toolbar'
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
