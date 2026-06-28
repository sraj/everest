import { useState, useEffect, useRef, useCallback } from 'react'
import type { Editor } from '@tiptap/react'

interface FindReplaceProps {
  editor: Editor
}

function textPositionToDocPos(editor: Editor, text: string, searchIndex: number, searchLen: number): { from: number; to: number } | null {
  let textPos = 0
  let foundStart = -1
  let foundEnd = -1

  editor.state.doc.descendants((node, pos) => {
    if (foundEnd >= 0) return false
    if (!node.isText) return

    const nodeText = node.text || ''
    for (let i = 0; i < nodeText.length; i++) {
      if (textPos >= searchIndex && foundStart < 0) {
        foundStart = pos + i
      }
      if (textPos >= searchIndex + searchLen - 1 && foundStart >= 0) {
        foundEnd = pos + i + 1
        return false
      }
      textPos++
    }
  })

  return foundStart >= 0 && foundEnd >= 0 ? { from: foundStart, to: foundEnd } : null
}

export function FindReplace({ editor }: FindReplaceProps) {
  const [visible, setVisible] = useState(false)
  const [query, setQuery] = useState('')
  const [replace, setReplace] = useState('')
  const [showReplace, setShowReplace] = useState(false)
  const [matchIndex, setMatchIndex] = useState(0)
  const [totalMatches, setTotalMatches] = useState(0)
  const [matches, setMatches] = useState<{ from: number; to: number }[]>([])
  const inputRef = useRef<HTMLInputElement>(null)

  const findMatches = useCallback((q: string) => {
    if (!q) {
      setMatches([])
      setTotalMatches(0)
      setMatchIndex(0)
      return
    }
    const text = editor.state.doc.textContent
    const results: { from: number; to: number }[] = []
    let idx = 0
    while ((idx = text.indexOf(q, idx)) >= 0) {
      const pos = textPositionToDocPos(editor, text, idx, q.length)
      if (pos) {
        results.push(pos)
        // highlight
        editor.chain().setTextSelection(pos).setHighlight({ color: '#fde047' }).run()
      }
      idx++
    }
    setMatches(results)
    setTotalMatches(results.length)
    if (results.length > 0) {
      setMatchIndex(0)
      editor.commands.setTextSelection(results[0])
    }
  }, [editor])

  useEffect(() => {
    findMatches(query)
  }, [query, findMatches])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
        e.preventDefault()
        setVisible(true)
        setTimeout(() => inputRef.current?.focus(), 0)
      }
      if (e.key === 'Escape') {
        setVisible(false)
        setQuery('')
        editor.commands.unsetHighlight()
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [editor])

  const navigateTo = (dir: 1 | -1) => {
    if (matches.length === 0) return
    const next = (matchIndex + dir + matches.length) % matches.length
    setMatchIndex(next)
    editor.commands.setTextSelection(matches[next])
  }

  const replaceOne = () => {
    if (matches.length === 0 || matchIndex >= matches.length) return
    const m = matches[matchIndex]
    editor.chain().setTextSelection(m).insertContent(replace).run()
    // Refind matches after content change
    setTimeout(() => findMatches(query), 0)
  }

  const replaceAll = () => {
    if (matches.length === 0) return
    // Replace all from last to first to preserve positions
    const sorted = [...matches].sort((a, b) => b.from - a.from)
    sorted.forEach((m) => {
      editor.chain().setTextSelection(m).insertContent(replace).run()
    })
    setTimeout(() => findMatches(query), 0)
  }

  const close = () => {
    setVisible(false)
    setQuery('')
    setReplace('')
    setShowReplace(false)
    editor.commands.unsetHighlight()
  }

  if (!visible) return null

  return (
    <div className="absolute top-2 right-2 z-30 flex items-center gap-1 rounded-lg border border-neutral-200 bg-white px-2 py-1.5 shadow-lg dark:border-neutral-700 dark:bg-neutral-900">
      <input
        ref={inputRef}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Find..."
        className="w-40 bg-transparent text-sm text-neutral-700 outline-none placeholder:text-neutral-400 dark:text-neutral-300 dark:placeholder:text-neutral-500"
      />
      <span className="text-[11px] tabular-nums text-neutral-400 dark:text-neutral-500 min-w-[4ch] text-right">
        {totalMatches > 0 ? `${matchIndex + 1} of ${totalMatches}` : query ? '0' : ''}
      </span>
      <button onClick={() => navigateTo(-1)} disabled={matches.length === 0} className="p-0.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 disabled:opacity-30 transition-colors" title="Previous match">
        <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m18 15-6-6-6 6"/></svg>
      </button>
      <button onClick={() => navigateTo(1)} disabled={matches.length === 0} className="p-0.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 disabled:opacity-30 transition-colors" title="Next match">
        <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m6 9 6 6 6-6"/></svg>
      </button>
      <button onClick={() => setShowReplace(!showReplace)} className={`p-0.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors ${showReplace ? 'text-blue-600 dark:text-blue-400' : 'text-neutral-400 dark:text-neutral-500'}`} title="Toggle replace">
        <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="17 1 21 5 17 9"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><polyline points="7 23 3 19 7 15"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/></svg>
      </button>
      {showReplace && (
        <>
          <input
            value={replace}
            onChange={(e) => setReplace(e.target.value)}
            placeholder="Replace..."
            className="w-32 bg-transparent text-sm text-neutral-700 outline-none placeholder:text-neutral-400 dark:text-neutral-300 dark:placeholder:text-neutral-500"
          />
          <button onClick={replaceOne} disabled={matches.length === 0} className="rounded px-2 py-0.5 text-[11px] font-medium bg-neutral-100 hover:bg-neutral-200 dark:bg-neutral-800 dark:hover:bg-neutral-700 disabled:opacity-30 transition-colors">Replace</button>
          <button onClick={replaceAll} disabled={matches.length === 0} className="rounded px-2 py-0.5 text-[11px] font-medium bg-neutral-100 hover:bg-neutral-200 dark:bg-neutral-800 dark:hover:bg-neutral-700 disabled:opacity-30 transition-colors">All</button>
        </>
      )}
      <button onClick={close} className="p-0.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 text-neutral-400 dark:text-neutral-500 transition-colors" title="Close (Esc)">
        <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
    </div>
  )
}
