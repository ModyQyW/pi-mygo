import type { ButtonHTMLAttributes } from "react";

import { cn } from "@/lib/cn";

type Variant = "primary" | "outline" | "ghost";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
};

const variants: Record<Variant, string> = {
  primary:
    "bg-coral text-ink border border-transparent hover:brightness-[0.96] active:brightness-[0.92]",
  outline:
    "border border-hairline-strong bg-surface text-ink hover:bg-sunken active:bg-sunken",
  ghost: "border border-transparent text-ink-2 hover:bg-sunken hover:text-ink",
};

/**
 * A control whose three variants are the only ones the design allows. Sizing is
 * fixed: 28px tall, 12px of horizontal padding, because a mixed toolbar with
 * mismatched heights reads as carelessness.
 */
export function Button({ variant = "outline", className, ...props }: ButtonProps) {
  return (
    <button
      type="button"
      className={cn(
        "inline-flex h-7 items-center justify-center gap-1.5 rounded-[var(--radius-chip)] px-3",
        "text-[13px] leading-none transition-colors duration-[var(--dur-hover)]",
        "disabled:pointer-events-none disabled:opacity-45",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}