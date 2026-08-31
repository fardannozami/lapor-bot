import { useState, useEffect } from "react";
import { Sparkles } from "lucide-react";

export function MotivationBanner() {
  const [quote, setQuote] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch("/api/motivation")
      .then((res) => res.json())
      .then((data) => {
        if (!cancelled && data.quote) setQuote(data.quote);
      })
      .catch(() => {
        if (!cancelled) setQuote(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!quote) return null;

  return (
    <div className="mb-6 px-5 py-4 glass glass-glow-volt rounded-2xl border border-system-volt/25 flex items-center gap-4 animate-fade-in">
      {/* Live match-ticker dot */}
      <span className="relative flex h-2.5 w-2.5 shrink-0" aria-hidden="true">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-system-volt opacity-75"></span>
        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-system-volt"></span>
      </span>
      <div className="min-w-0">
        <p className="text-[10px] font-mono font-bold uppercase tracking-widest text-system-volt flex items-center gap-1.5">
          <Sparkles size={11} className="text-system-gold" />
          Pep Talk // Daily Boost
        </p>
        <p className="text-sm text-gray-300 font-medium italic mt-1">
          &ldquo;{quote}&rdquo;
        </p>
      </div>
    </div>
  );
}
