import {
  IconArrowBackUp,
  IconArrowForwardUp,
  IconBold,
  IconItalic,
  IconUnderline,
  IconStrikethrough,
  IconSuperscript,
  IconSubscript,
  IconPalette,
  IconLink,
  IconPhotoUp,
  IconH1,
  IconH2,
  IconH3,
  IconH4,
  IconH5,
  IconH6,
  IconPilcrow,
  IconList,
  IconListNumbers,
  IconListCheck,
  IconBlockquote,
  IconCode,
  IconAlignLeft,
  IconAlignCenter,
  IconAlignRight,
  IconMinus,
  IconClearFormatting,
  IconPrinter,
  IconChevronDown,
  IconChevronLeft,
  IconFileText,
  IconTable,
  IconRowInsertTop,
  IconColumns,
  IconIndentIncrease,
  IconIndentDecrease,
} from '@tabler/icons-react'

function icon(Icon: React.FC<{ size?: number | string; stroke?: number }>) {
  return <Icon size={20} stroke={1.5} />
}

export function Undo() { return icon(IconArrowBackUp) }
export function Redo() { return icon(IconArrowForwardUp) }
export function Bold() { return icon(IconBold) }
export function Italic() { return icon(IconItalic) }
export function Underline() { return icon(IconUnderline) }
export function Strikethrough() { return icon(IconStrikethrough) }
export function Color() { return icon(IconPalette) }
export function Image() { return icon(IconPhotoUp) }
export function BulletList() { return icon(IconList) }
export function OrderedList() { return icon(IconListNumbers) }
export function TaskList() { return icon(IconListCheck) }
export function Blockquote() { return icon(IconBlockquote) }
export function Code() { return icon(IconCode) }
export function HorizontalRule() { return icon(IconMinus) }
export function ClearFormatting() { return icon(IconClearFormatting) }
export function PageIcon() { return icon(IconFileText) }
export function SuperscriptIcon() { return icon(IconSuperscript) }
export function SubscriptIcon() { return icon(IconSubscript) }
export function TableIconComponent() { return icon(IconTable) }
export function AddRowBefore() { return icon(IconRowInsertTop) }
export function AddColumnBefore() { return icon(IconColumns) }
export function IndentMore() { return icon(IconIndentIncrease) }
export function IndentLess() { return icon(IconIndentDecrease) }

export { IconChevronDown as ChevronDown, IconChevronLeft as ChevronLeft }
export { IconAlignLeft as AlignLeft, IconAlignCenter as AlignCenter, IconAlignRight as AlignRight }
export { IconH1 as Heading1, IconH2 as Heading2, IconH3 as Heading3, IconH4 as Heading4, IconH5 as Heading5, IconH6 as Heading6 }
export { IconPilcrow as Paragraph, IconLink as Link, IconPrinter as Printer }
