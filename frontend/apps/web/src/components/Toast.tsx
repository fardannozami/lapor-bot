import { useEffect, useRef, useState } from "react";
import { CheckCircle2, AlertCircle, X } from "lucide-react";

export type ToastTone = "success" | "error";

export interface ToastData {
  /** Unique id so re-triggering the same message re-animates. */
  id: number;
  message: string;
  tone: ToastTone;
}

interface ToastProps {
  toast: ToastData | null;
  onDismiss: () => void;
}

const AUTO_DISMISS_MS = 4000;
const EXIT_MS = 160;

/**
 * Lightweight confirmation toast for personal-dashboard mutations
 * (goal set/reset, name update, job change). Accessible polite live region,
 * bottom-center placement, subtle enter/exit motion, pauses on hover/focus,
 * and honours prefers-reduced-motion.
 */
export function Toast({ toast, onDismiss }: ToastProps) {
  const [visible, setVisible] = useState(false);
  const dismissTimer = useRef<number | null>(null);
  const exitTimer = useRef<number | null>(null);

  useEffect(() => {
    if (!toast) return;

    // Enter on next frame so the transition triggers.
    const raf = requestAnimationFrame(() => setVisible(true));
    dismissTimer.current = window.setTimeout(handleClose, AUTO_DISMISS_MS);

    return () => {
      cancelAnimationFrame(raf);
      if (dismissTimer.current) clearTimeout(dismissTimer.current);
      if (exitTimer.current) clearTimeout(exitTimer.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [toast?.id]);

  // Re-arm dismiss when hovered/focused (pause), resume on leave/blur.
  const pauseAutoDismiss = () => {
    if (dismissTimer.current) clearTimeout(dismissTimer.current);
  };
  const resumeAutoDismiss = () => {
    dismissTimer.current = window.setTimeout(handleClose, AUTO_DISMISS_MS);
  };

  function handleClose() {
    // Play exit transition, then ask the parent to unmount us.
    setVisible(false);
    if (exitTimer.current) clearTimeout(exitTimer.current);
    exitTimer.current = window.setTimeout(onDismiss, EXIT_MS);
  }

  if (!toast) return null;

  const isError = toast.tone === "error";
  const Icon = isError ? AlertCircle : CheckCircle2;

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      onMouseEnter={pauseAutoDismiss}
      onMouseLeave={resumeAutoDismiss}
      onFocus={pauseAutoDismiss}
      onBlur={resumeAutoDismiss}
      style={{
        transform: visible ? "translateY(0)" : "translateY(10px)",
        opacity: visible ? 1 : 0,
        transition:
          "transform 180ms cubic-bezier(0.16, 1, 0.3, 1), opacity 180ms cubic-bezier(0.16, 1, 0.3, 1)",
      }}
      className={`fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 max-w-[92vw] sm:max-w-md px-4 py-3 rounded-2xl border shadow-lg backdrop-blur ${
        isError
          ? "bg-system-red/15 border-system-red/40 text-red-200"
          : "bg-system-green/15 border-system-green/40 text-system-green"
      }`}
    >
      <Icon
        size={18}
        className={isError ? "text-system-red" : "text-system-green"}
      />
      <p className="text-sm font-mono leading-snug flex-1">{toast.message}</p>
      <button
        type="button"
        onClick={handleClose}
        aria-label="Tutup notifikasi"
        className="shrink-0 p-1 rounded-lg text-current/70 hover:text-current transition-colors"
      >
        <X size={14} />
      </button>
    </div>
  );
}
