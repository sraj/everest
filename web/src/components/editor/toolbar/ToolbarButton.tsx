import { forwardRef } from 'react'
import type { ButtonHTMLAttributes, ForwardedRef } from 'react'

interface ToolbarButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean
}

function ToolbarButtonInner({ active, className = '', children, ...props }: ToolbarButtonProps, ref: ForwardedRef<HTMLButtonElement>) {
  return (
    <button
      ref={ref}
      data-state={active ? 'on' : 'off'}
      className={`inline-flex items-center justify-center size-8 rounded-md transition-colors duration-150 disabled:opacity-40 disabled:cursor-not-allowed
        data-[state=on]:bg-neutral-200 data-[state=on]:text-neutral-900 dark:data-[state=on]:bg-neutral-700 dark:data-[state=on]:text-white
        text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-800 dark:hover:text-white
        ${className}`}
      {...props}
    >
      {children}
    </button>
  )
}

export const ToolbarButton = forwardRef(ToolbarButtonInner)

export function ToolbarDivider() {
  return <div className="mx-1 h-6 w-px bg-neutral-200 dark:bg-neutral-700" />
}
