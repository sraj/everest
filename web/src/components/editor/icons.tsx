import {
  Undo2,
  Redo2,
  Bold,
  Italic,
  Underline,
  Strikethrough,
  Superscript,
  Subscript,
  Palette,
  Link,
  ImageUp,
  Heading1,
  Heading2,
  Heading3,
  Heading4,
  Heading5,
  Heading6,
  Pilcrow,
  List,
  ListOrdered,
  ListChecks,
  Quote,
  Braces,
  AlignLeft,
  AlignCenter,
  AlignRight,
  Minus,
  RemoveFormatting,
  Printer,
  ChevronDown,
  ChevronLeft,
  FileText,
} from 'lucide-react'

export function Undo() { return <Undo2 className="size-4" /> }
export function Redo() { return <Redo2 className="size-4" /> }
export { Bold, Italic, Underline, Strikethrough }
export function Color() { return <Palette className="size-4" /> }
export { Link }
export function Image() { return <ImageUp className="size-4" /> }
export { Heading1, Heading2, Heading3, Heading4, Heading5, Heading6 }
export function Paragraph() { return <Pilcrow className="size-4" /> }
export function BulletList() { return <List className="size-4" /> }
export function OrderedList() { return <ListOrdered className="size-4" /> }
export function TaskList() { return <ListChecks className="size-4" /> }
export function Blockquote() { return <Quote className="size-4" /> }
export function Code() { return <Braces className="size-4" /> }
export { AlignLeft, AlignCenter, AlignRight }
export { Printer }
export { ChevronDown, ChevronLeft }
export function HorizontalRule() { return <Minus className="size-4" /> }
export function ClearFormatting() { return <RemoveFormatting className="size-4" /> }
export function PageIcon() { return <FileText className="size-4" /> }

export function SuperscriptIcon() { return <Superscript className="size-4" /> }
export function SubscriptIcon() { return <Subscript className="size-4" /> }
