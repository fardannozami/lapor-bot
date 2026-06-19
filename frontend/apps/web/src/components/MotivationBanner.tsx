import { useState, useEffect } from 'react';
import { Sparkles } from 'lucide-react';

export function MotivationBanner() {
  const [quote, setQuote] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch('/api/motivation')
      .then((res) => res.json())
      .then((data) => {
        if (!cancelled && data.quote) setQuote(data.quote);
      })
      .catch(() => {
        if (!cancelled) setQuote(null);
      });
    return () => { cancelled = true; };
  }, []);

  if (!quote) return null;

  return (
    <div className="mb-6 px-5 py-4 glass rounded-2xl border border-system-green/20 flex items-center gap-3">
      <Sparkles className="text-system-gold shrink-0" size={18} />
      <p className="text-sm text-gray-300 font-medium italic">
        &ldquo;{quote}&rdquo;
      </p>
    </div>
  );
}
