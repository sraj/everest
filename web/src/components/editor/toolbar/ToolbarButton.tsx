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
      className={`inline-flex items-center justify-center size-8 rounded transition-colors duration-100 disabled:opacity-30 disabled:cursor-not-allowed
        data-[state=on]:bg-neutral-300/60 data-[state=on]:text-neutral-900 dark:data-[state=on]:bg-neutral-600 dark:data-[state=on]:text-white
        text-neutral-600 hover:bg-neutral-200/60 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-700/60 dark:hover:text-white
        ${className}`}
      {...props}
    >
      {children}
    </button>
  )
}

export const ToolbarButton = forwardRef(ToolbarButtonInner)

export function ToolbarDivider() {
  return <div className="mx-0.5 h-5 w-px bg-neutral-300 dark:bg-neutral-600" />
}
