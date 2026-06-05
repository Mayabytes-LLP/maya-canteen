import { Button, type ButtonProps } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { CheckIcon, ClipboardIcon } from "lucide-react";
import { useEffect, useState } from "react";
interface CopyButtonProps extends ButtonProps {
  value: string;
  timeout?: number;
  }

export async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
  } catch (err) {
    console.error("Failed to copy to clipboard", err);
  }
}

export function CopyButton({
  value,
  className,
  timeout = 3000,
  variant = "ghost",
  ...props
}: CopyButtonProps) {
  const [hasCopied, setHasCopied] = useState(false);

  // change icon back to clipboard after timeout
  useEffect(() => {
    if (!hasCopied) return;
    const id = setTimeout(() => {
      setHasCopied(false);
    }, timeout);
    return () => clearTimeout(id);
  }, [hasCopied, timeout]);

  return (
    <Button
      size="icon"
      variant={variant}
      className={cn(
        "relative z-10 h-6 w-6 text-zinc-50 hover:bg-zinc-700 hover:text-zinc-50 [&_svg]:h-3 [&_svg]:w-3",
        className,
        hasCopied && "text-green-500",
      )}
      onClick={() => {
        copyToClipboard(value);
        setHasCopied(true);
      }}
      {...props}
    >
      <span className="sr-only">Copy</span>
      {hasCopied ? <CheckIcon /> : <ClipboardIcon />}
    </Button>
  );
}
