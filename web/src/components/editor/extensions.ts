import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Color from '@tiptap/extension-color'
import { TextStyle } from '@tiptap/extension-text-style'
import TextAlign from '@tiptap/extension-text-align'
import Highlight from '@tiptap/extension-highlight'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Superscript from '@tiptap/extension-superscript'
import Subscript from '@tiptap/extension-subscript'
import Placeholder from '@tiptap/extension-placeholder'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import { CharacterCount } from '@tiptap/extension-character-count'
import { PaginationPlus, PAGE_SIZES } from 'tiptap-pagination-plus'
import type { PageSize } from 'tiptap-pagination-plus'
export { PAGE_SIZES } from 'tiptap-pagination-plus'
import { FontFamily } from '@tiptap/extension-text-style/font-family'
import { FontSize } from '@tiptap/extension-text-style/font-size'
import { LineHeight } from '@tiptap/extension-text-style/line-height'

function makeSize(h: number, w: number, mt = 95, mb = 95, ml = 76, mr = 76): PageSize {
  return { pageHeight: h, pageWidth: w, marginTop: mt, marginBottom: mb, marginLeft: ml, marginRight: mr }
}

export const CUSTOM_PAGE_SIZES = {
  STATEMENT: makeSize(529, 816, 96, 96, 96, 96),
  EXECUTIVE: makeSize(1009, 695, 96, 96, 96, 96),
  B4: makeSize(1334, 945),
  B5: makeSize(945, 665),
}

export const ALL_PAGE_SIZES = { ...PAGE_SIZES, ...CUSTOM_PAGE_SIZES }
export type PageSizeKey = keyof typeof ALL_PAGE_SIZES

export const PAGE_SIZE_LABELS: Record<PageSizeKey, string> = {
  A4: 'A4',
  A3: 'A3',
  A5: 'A5',
  LETTER: 'Letter',
  LEGAL: 'Legal',
  TABLOID: 'Tabloid',
  STATEMENT: 'Statement',
  EXECUTIVE: 'Executive',
  B4: 'B4',
  B5: 'B5',
}

export const PAGE_SIZE_ORDER: PageSizeKey[] = [
  'A4', 'A3', 'A5', 'LETTER', 'LEGAL', 'TABLOID',
  'STATEMENT', 'EXECUTIVE', 'B4', 'B5',
]

export const FONT_FAMILIES = [
  'Arial',
  'Helvetica',
  'Times New Roman',
  'Georgia',
  'Courier New',
  'Verdana',
  'Trebuchet MS',
  'Comic Sans MS',
]

export const FONT_SIZES = [8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 26, 28, 36, 42, 48, 60, 72]

export const LINE_HEIGHTS = ['1', '1.15', '1.5', '2', '2.5', '3']

export function createExtensions(pageSize: PageSizeKey) {
  return [
    StarterKit.configure({
      link: false,
    }),
    Underline,
    TextStyle,
    FontFamily,
    FontSize,
    LineHeight,
    Color,
    Highlight.configure({ multicolor: true }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'tiptap-link',
      },
    }),
    Image.configure({
      inline: false,
      allowBase64: true,
    }),
    TaskList.configure({
      HTMLAttributes: {
        class: 'tiptap-task-list',
      },
    }),
    TaskItem.configure({
      nested: true,
      HTMLAttributes: {
        class: 'tiptap-task-item',
      },
    }),
    Superscript,
    Subscript,
    TextAlign.configure({
      types: ['heading', 'paragraph'],
    }),
    Placeholder.configure({
      placeholder: 'Start typing...',
    }),
    Table.configure({
      resizable: true,
    }),
    TableRow,
    TableCell,
    TableHeader,
    CharacterCount,
    PaginationPlus.configure({
      pageGap: 24,
      contentMarginTop: 0,
      contentMarginBottom: 0,
      ...ALL_PAGE_SIZES[pageSize],
      marginTop: 76,
      marginBottom: 76,
      pageBreakBackground: '#f5f5f5',
    }),
  ]
}
