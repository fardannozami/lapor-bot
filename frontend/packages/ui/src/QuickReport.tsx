import { useState, useRef, type FormEvent } from "react";
import {
  Send,
  Loader2,
  Dumbbell,
  Footprints,
  Bike,
  Timer,
  Sparkles,
  ChevronDown,
  ScrollText,
  Activity,
} from "lucide-react";
import type { ReportResult } from "@lapor-bot/shared";

export interface QuickReportCallbacks {
  createReport(message: string): Promise<ReportResult>;
  createSidequestReport(activities: string[]): Promise<ReportResult>;
}

interface QuickReportProps {
  /** Called after a successful report to refresh user data. */
  onReported: () => Promise<void>;
  /** Repository-mapped callbacks for API calls. */
  callbacks: QuickReportCallbacks;
  /** Current report count today (0-3). */
  dailyCount: number;
  /** Maximum reports allowed per day. */
  maxDaily: number;
}

const QUICK_CHIPS = [
  { label: "Gym", icon: Dumbbell, message: "Gym hari ini" },
  { label: "Lari", icon: Footprints, message: "Lari" },
  { label: "Sepeda", icon: Bike, message: "Sepeda" },
  { label: "Olahraga", icon: Activity, message: "Olahraga" },
];

const SIDEQUEST_QUICK = [
  { label: "Jalan Kaki", value: "jalan kaki 4000" },
  { label: "Push Up", value: "pushup 15" },
  { label: "Squat", value: "squat 20" },
  { label: "Sepeda", value: "sepeda 5 km" },
];

export function QuickReport({
  onReported,
  callbacks,
  dailyCount,
  maxDaily,
}: QuickReportProps) {
  const [message, setMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState<{
    text: string;
    tone: "success" | "error";
  } | null>(null);
  const [showSidequest, setShowSidequest] = useState(false);
  const [sidequestInput, setSidequestInput] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const sidequestRef = useRef<HTMLInputElement>(null);
  const feedbackTimer = useRef<number | null>(null);

  const capped = dailyCount >= maxDaily;

  const showFeedback = (text: string, tone: "success" | "error") => {
    setFeedback({ text, tone });
    if (feedbackTimer.current) clearTimeout(feedbackTimer.current);
    feedbackTimer.current = window.setTimeout(() => setFeedback(null), 4000);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (capped) return;
    const text = message.trim();
    setSubmitting(true);
    setFeedback(null);
    try {
      const result = await callbacks.createReport(text || "Olahraga");
      showFeedback(result.message, "success");
      setMessage("");
      inputRef.current?.focus();
      await onReported();
    } catch (err: any) {
      showFeedback(err?.message || "Gagal mengirim laporan", "error");
    } finally {
      setSubmitting(false);
    }
  };

  const handleSidequestSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const value = sidequestInput.trim();
    if (!value) return;
    setSubmitting(true);
    setFeedback(null);
    try {
      const result = await callbacks.createSidequestReport([value]);
      showFeedback(result.message, "success");
      setSidequestInput("");
      setShowSidequest(false);
      await onReported();
    } catch (err: any) {
      showFeedback(
        err?.message || "Gagal mengirim sidequest",
        "error",
      );
    } finally {
      setSubmitting(false);
    }
  };

  const handleChipClick = (chipMessage: string) => {
    setMessage(chipMessage);
    inputRef.current?.focus();
  };

  return (
    <section className="glass rounded-3xl p-4 sm:p-6 glass-glow-green mb-6">
      <div className="flex items-center justify-between gap-3 mb-4">
        <div>
          <h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
            <Sparkles className="text-system-gold" size={18} />
            Lapor Hari Ini
          </h3>
          <p className="text-xs text-gray-500 font-mono mt-1">
            {capped
              ? `Kuota harian penuh (${dailyCount}/${maxDaily})`
              : `Laporan ke-${dailyCount + 1} hari ini (max ${maxDaily}x)`}
          </p>
        </div>
        {dailyCount > 0 && (
          <span
            className={`text-sm font-bold font-orbitron ${
              capped ? "text-system-red" : "text-system-green"
            }`}
          >
            {dailyCount}/{maxDaily}
          </span>
        )}
      </div>

      {/* Main report form */}
      <form onSubmit={handleSubmit} className="mb-3">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <input
              ref={inputRef}
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder={
                capped
                  ? "Kuota laporan harian penuh"
                  : "Apa aktivitasmu hari ini?"
              }
              disabled={capped || submitting}
              className="w-full px-4 py-3 rounded-2xl bg-gray-950 border border-gray-800 text-white text-base sm:text-sm font-mono placeholder:text-gray-600 focus:outline-none focus:border-system-green/50 focus:ring-1 focus:ring-system-green/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            />
          </div>
          <button
            type="submit"
            disabled={capped || submitting}
            className="shrink-0 flex items-center gap-2 px-5 py-3 rounded-2xl bg-system-green hover:bg-system-green/90 text-gray-950 font-bold font-orbitron text-sm transition-all disabled:opacity-40 disabled:cursor-not-allowed min-h-[44px]"
          >
            {submitting ? (
              <Loader2 className="animate-spin" size={16} />
            ) : (
              <Send size={16} />
            )}
            Lapor
          </button>
        </div>
      </form>

      {/* Quick chips */}
      {!capped && (
        <div className="flex flex-wrap gap-2 mb-3">
          {QUICK_CHIPS.map((chip) => {
            const Icon = chip.icon;
            return (
              <button
                key={chip.label}
                type="button"
                disabled={submitting}
                onClick={() => handleChipClick(chip.message)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-gray-950/70 border border-gray-800 text-xs font-mono text-gray-300 hover:text-white hover:border-gray-600 transition-colors disabled:opacity-50 min-h-[36px]"
              >
                <Icon size={12} />
                {chip.label}
              </button>
            );
          })}
        </div>
      )}

      {/* Sidequest expandable section */}
      <div>
        <button
          type="button"
          onClick={() => {
            setShowSidequest((v) => !v);
            if (!showSidequest) {
              setTimeout(() => sidequestRef.current?.focus(), 100);
            }
          }}
          className="flex items-center gap-2 text-xs font-mono text-gray-400 hover:text-system-green transition-colors min-h-[32px]"
        >
          <ScrollText size={14} />
          Lapor Sidequest
          <ChevronDown
            size={12}
            className={`transition-transform ${showSidequest ? "rotate-180" : ""}`}
          />
        </button>

        {showSidequest && (
          <div className="mt-3 p-4 rounded-2xl bg-gray-950/50 border border-gray-800">
            <p className="text-[10px] text-gray-500 font-mono mb-3">
              Format: &lt;kegiatan&gt; &lt;jumlah&gt;. Contoh: pushup 15, lari 10
              km, plank 30 detik
            </p>

            <div className="flex flex-wrap gap-1.5 mb-3">
              {SIDEQUEST_QUICK.map((sq) => (
                <button
                  key={sq.label}
                  type="button"
                  disabled={submitting}
                  onClick={() => setSidequestInput(sq.value)}
                  className="px-2.5 py-1 rounded-lg bg-gray-900 border border-gray-800 text-[10px] font-mono text-gray-400 hover:text-white hover:border-gray-600 transition-colors disabled:opacity-50"
                >
                  {sq.label}
                </button>
              ))}
            </div>

            <form onSubmit={handleSidequestSubmit} className="flex gap-2">
              <input
                ref={sidequestRef}
                type="text"
                value={sidequestInput}
                onChange={(e) => setSidequestInput(e.target.value)}
                placeholder="pushup 15"
                disabled={submitting}
                className="flex-1 px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-base sm:text-sm font-mono text-white placeholder:text-gray-600 focus:outline-none focus:border-system-green/50 transition-all disabled:opacity-50"
              />
              <button
                type="submit"
                disabled={submitting || !sidequestInput.trim()}
                className="shrink-0 flex items-center gap-1.5 px-4 py-2 rounded-xl bg-system-green/15 hover:bg-system-green/25 border border-system-green/30 text-system-green font-mono text-xs transition-colors disabled:opacity-40 disabled:cursor-not-allowed min-h-[40px]"
              >
                {submitting ? (
                  <Loader2 className="animate-spin" size={12} />
                ) : (
                  <Timer size={12} />
                )}
                Kirim
              </button>
            </form>
          </div>
        )}
      </div>

      {/* Feedback toast */}
      {feedback && (
        <div
          className={`mt-3 p-3 rounded-xl border text-xs font-mono ${
            feedback.tone === "success"
              ? "border-system-green/30 bg-system-green/10 text-system-green"
              : "border-system-red/30 bg-system-red/10 text-system-red"
          }`}
        >
          {feedback.text}
        </div>
      )}
    </section>
  );
}
