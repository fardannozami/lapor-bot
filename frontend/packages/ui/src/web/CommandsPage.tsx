import { useEffect, useRef, useState } from "react";
import {
  BookOpen,
  CalendarClock,
  Check,
  ChevronLeft,
  CircleHelp,
  Copy,
  Dumbbell,
  Footprints,
  Globe,
  Link2,
  MessageCircle,
  Send,
  Sparkles,
  Target,
  Terminal,
  Trophy,
  UserCircle2,
  XCircle,
  type LucideIcon,
} from "lucide-react";

/* -------------------------------------------------------------------------- */
/* Data — source of truth: internal/app/usecase/get_help_usecase.go            */
/* -------------------------------------------------------------------------- */

interface WhatsAppCommand {
  /** Primary syntax, e.g. "/lapor jalan kaki [jumlah]". */
  syntax: string;
  /** Equivalent forms, e.g. the "#" prefix variant. */
  aliases?: string[];
  description: string;
  examples?: string[];
  note?: string;
}

interface CommandCategory {
  id: string;
  icon: LucideIcon;
  title: string;
  summary: string;
  commands: WhatsAppCommand[];
}

const WHATSAPP_CATEGORIES: CommandCategory[] = [
  {
    id: "lapor",
    icon: Dumbbell,
    title: "Melaporkan Aktivitas",
    summary:
      "Cara utama mencatat workout. Laporan valid menambah streak mingguan, XP, dan total hari aktif.",
    commands: [
      {
        syntax: "/lapor",
        aliases: ["#lapor"],
        description:
          "Laporkan workout atau aktivitas harianmu. Posisi command fleksibel — bisa di awal, tengah, atau akhir kalimat.",
        examples: ["/lapor", "/lapor Push Day", "gym pagi ini /lapor"],
        note: "Max 3x laporan per hari. Laporan ke-2 dan ke-3 tetap dihitung 1 hari, tapi XP dibagi 2. Kamu punya 1 Streak Freeze gratis per season, +1 lagi setiap mencapai 4 minggu streak.",
      },
      {
        syntax: "/lapor-kemarin",
        aliases: ["#lapor-kemarin"],
        description:
          "Laporan khusus untuk hari kemarin — berguna kalau lupa lapor. Sama seperti /lapor, XP dibagi 2.",
        examples: ["/lapor-kemarin lari 5K", "/lapor-kemarin gym"],
        note: "Max 3x per hari, berbagi slot dengan laporan hari ini.",
      },
      {
        syntax: "/lapor jalan kaki [jumlah]",
        aliases: ["#lapor jalan kaki [jumlah]"],
        description:
          "Catat langkah harian sebagai laporan utama — masuk streak, XP, leaderboard, dan klasemen langkah di web.",
        examples: [
          "/lapor jalan kaki 5000",
          "/lapor jalan kaki 12.500",
          "/lapor jalan kaki 8rb",
          "walking 6500 steps /lapor",
        ],
        note: "Angka langkah WAJIB (min 100, max 250.000) — tanpa angka valid laporan tidak tercatat. Format bebas: 12.500, 12,500, 8rb, 8k, atau nempel 12500langkah. Durasi seperti `jalan 30 menit` dianggap laporan biasa.",
      },
    ],
  },
  {
    id: "sidequest",
    icon: Sparkles,
    title: "Side Quest Harian",
    summary:
      "Quest bonus harian untuk hunter yang sudah punya job — easy, medium, dan hard dengan XP bonus per difficulty.",
    commands: [
      {
        syntax: "/lapor sidequest",
        aliases: ["#lapor sidequest"],
        description:
          "Lihat daftar side quest easy, medium, dan hard hari ini untuk profilmu.",
        note: "Easy: jalan kaki min 4.000 langkah atau sepeda 5 km. Medium/hard: latihan ringan yang naik sedikit sesuai level job.",
      },
      {
        syntax: "/lapor sidequest [kegiatan] [jumlah]",
        aliases: ["#lapor sidequest [kegiatan] [jumlah]"],
        description:
          "Lapor penyelesaian side quest. Nama kegiatan harus sama persis dengan yang tertera di daftar quest, dan target harus tercapai.",
        examples: [
          "/lapor sidequest jalan kaki 4000",
          "/lapor sidequest sepeda 5 km",
          "/lapor sidequest chair squat 18",
        ],
        note: "Slot side quest terpisah (3x/hari) dari laporan utama. Langkah dari side quest easy ikut masuk akumulasi jalan kaki di web.",
      },
    ],
  },
  {
    id: "cancel",
    icon: XCircle,
    title: "Membatalkan Laporan",
    summary:
      "Salah input? Batalkan laporanmu — hanya bisa pada hari yang sama dengan laporan.",
    commands: [
      {
        syntax: "/cancel",
        aliases: ["#cancel"],
        description:
          "Batalkan laporan utama terakhir hari ini. Kalau ada 2–3 laporan, hanya yang paling akhir yang dihapus.",
      },
      {
        syntax: "/cancel-all",
        aliases: ["#cancel-all"],
        description:
          "Hapus semua laporan utama hari ini dan hitung ulang progresmu.",
      },
      {
        syntax: "/cancel sidequest",
        aliases: ["#cancel sidequest"],
        description: "Batalkan side quest terakhir hari ini.",
      },
      {
        syntax: "/cancel-all sidequest",
        aliases: ["#cancel-all sidequest"],
        description: "Hapus semua side quest hari ini.",
      },
    ],
  },
  {
    id: "strava",
    icon: Link2,
    title: "Integrasi Strava",
    summary:
      "Hubungkan Strava sekali, aktivitas larimu otomatis terlapor tanpa command.",
    commands: [
      {
        syntax: "/strava",
        description:
          "Dapatkan link otorisasi Strava. Kalau dikirim di grup, bot akan mengarahkanmu ke chat private demi keamanan.",
        note: "Setelah terhubung, aktivitas Strava otomatis dilaporkan bot.",
      },
      {
        syntax: "/disconnect-strava",
        aliases: ["/disconnectstrava"],
        description: "Putuskan koneksi akun Strava dari profilmu.",
      },
    ],
  },
  {
    id: "bantuan",
    icon: CircleHelp,
    title: "Bantuan",
    summary: "Panduan langsung dari bot di dalam chat.",
    commands: [
      {
        syntax: "/help",
        aliases: ["#help"],
        description: "Tampilkan list command ringkas.",
      },
      {
        syntax: "/tutorial",
        aliases: ["#tutorial"],
        description:
          "Tampilkan panduan lengkap penggunaan bot, termasuk attribute RPG dan flow side quest.",
      },
    ],
  },
];

interface WebsiteStep {
  icon: LucideIcon;
  title: string;
  description: string;
}

const WEBSITE_STEPS: WebsiteStep[] = [
  {
    icon: Trophy,
    title: "Buka Leaderboard",
    description:
      "Klasemen season, ranking, streak, dan klasemen langkah bisa dilihat siapa saja — tanpa perlu login.",
  },
  {
    icon: UserCircle2,
    title: "Masuk dengan Nomor WhatsApp",
    description:
      "Klik tombol Masuk di kanan atas, lalu gunakan nomor WhatsApp yang kamu pakai untuk lapor di grup.",
  },
  {
    icon: BookOpen,
    title: "Lengkapi Profil & Pilih Job",
    description:
      "Atur nama tampilan dan pilih job RPG-mu. Job menentukan attribute yang naik dan isi side quest harianmu.",
  },
  {
    icon: Send,
    title: "Lapor Langsung dari Web",
    description:
      "Panel Lapor Hari Ini di halaman personal bekerja persis seperti /lapor — termasuk lapor sidequest. Berbagi kuota 3x/hari dengan WhatsApp.",
  },
  {
    icon: Target,
    title: "Atur Goals Mingguan",
    description:
      "Tentukan target aktivitas mingguanmu. Saat goal tercapai, bot mengirim notifikasi selebrasi ke grup WhatsApp.",
  },
  {
    icon: Link2,
    title: "Hubungkan Strava",
    description:
      "Dari halaman personal, hubungkan Strava agar aktivitas lari/sepedamu terlapor otomatis tanpa perlu command.",
  },
];

const WEBSITE_URL = "https://lapor-bot.web.id/";

/* -------------------------------------------------------------------------- */
/* Small building blocks                                                       */
/* -------------------------------------------------------------------------- */

/** kbd-style command chip that copies itself on click (shadcn pattern). */
function SyntaxChip({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const timer = useRef<number | null>(null);

  useEffect(
    () => () => {
      if (timer.current) window.clearTimeout(timer.current);
    },
    [],
  );

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // Clipboard API blocked (non-secure context / permission) — fall back
      // to the legacy selection path so the feature still works everywhere.
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      try {
        document.execCommand("copy");
      } catch {
        // Give up silently — the chip is still readable text.
      }
      document.body.removeChild(ta);
    }
    setCopied(true);
    if (timer.current) window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setCopied(false), 1600);
  };

  return (
    <button
      type="button"
      onClick={copy}
      title={copied ? "Tersalin!" : `Salin ${text}`}
      aria-label={`Salin command ${text}`}
      className="group/chip inline-flex max-w-full items-center gap-1.5 min-h-[32px] rounded-lg border border-system-green/30 bg-system-green/10 px-2.5 font-mono text-xs font-semibold text-system-green shadow-[inset_0_0_10px_rgba(34,197,94,0.08)] hover:bg-system-green/20 hover:border-system-green/50 transition-colors cursor-pointer"
    >
      <span className="truncate">{text}</span>
      {copied ? (
        <Check size={12} className="shrink-0 text-system-volt" />
      ) : (
        <Copy
          size={12}
          className="shrink-0 opacity-40 group-hover/chip:opacity-100 transition-opacity"
        />
      )}
    </button>
  );
}

function CommandCard({ command }: { command: WhatsAppCommand }) {
  return (
    <article className="glass rounded-2xl p-4 sm:p-5 border border-gray-800 hover:border-gray-700 transition-colors">
      <div className="flex flex-wrap items-center gap-2">
        <SyntaxChip text={command.syntax} />
        {command.aliases?.map((alias) => (
          <SyntaxChip key={alias} text={alias} />
        ))}
      </div>
      <p className="mt-3 text-sm text-gray-300 leading-relaxed">
        {command.description}
      </p>
      {command.examples && command.examples.length > 0 && (
        <div className="mt-3 rounded-xl bg-gray-950/70 border border-gray-800/70 px-3 py-2.5">
          <p className="text-[10px] font-mono uppercase tracking-widest text-gray-600 mb-1.5">
            Contoh
          </p>
          <ul className="space-y-1">
            {command.examples.map((example) => (
              <li
                key={example}
                className="font-mono text-xs text-gray-400 flex items-center gap-2"
              >
                <span className="text-system-green/60 select-none">›</span>
                {example}
              </li>
            ))}
          </ul>
        </div>
      )}
      {command.note && (
        <p className="mt-3 text-[11px] font-mono leading-relaxed text-gray-500 border-l-2 border-system-gold/40 pl-3">
          {command.note}
        </p>
      )}
    </article>
  );
}

function CategorySection({
  category,
  index,
}: {
  category: CommandCategory;
  index: number;
}) {
  const Icon = category.icon;
  return (
    <section
      id={`command-${category.id}`}
      aria-labelledby={`command-${category.id}-heading`}
      className="scroll-mt-24"
    >
      <header className="mb-4">
        <p className="font-orbitron text-[10px] uppercase tracking-[0.25em] text-system-green/70">
          {String(index + 1).padStart(2, "0")} / {category.title}
        </p>
        <h3
          id={`command-${category.id}-heading`}
          className="mt-1.5 text-base sm:text-lg font-bold font-orbitron text-white flex items-center gap-2"
        >
          <Icon size={18} className="text-system-green shrink-0" />
          {category.title}
        </h3>
        <p className="mt-1 text-xs text-gray-500 font-mono leading-relaxed max-w-2xl">
          {category.summary}
        </p>
      </header>
      <div className="grid gap-4 lg:grid-cols-2">
        {category.commands.map((command) => (
          <CommandCard key={command.syntax} command={command} />
        ))}
      </div>
    </section>
  );
}

function WhatsAppGuide() {
  return (
    <div className="space-y-10">
      {/* Quick start strip */}
      <div className="glass glass-glow-green rounded-2xl p-4 sm:p-5 border border-system-green/20">
        <p className="font-orbitron text-xs font-bold uppercase tracking-widest text-system-green flex items-center gap-2">
          <Terminal size={14} />
          Quick Start
        </p>
        <ol className="mt-3 grid gap-2 sm:grid-cols-3 text-xs font-mono text-gray-300">
          <li className="flex items-start gap-2">
            <span className="text-system-green font-bold shrink-0">1.</span>
            Kirim <SyntaxChip text="/lapor" /> di grup WhatsApp
          </li>
          <li className="flex items-start gap-2">
            <span className="text-system-green font-bold shrink-0">2.</span>
            Bot membalas & progresmu tercatat
          </li>
          <li className="flex items-start gap-2">
            <span className="text-system-green font-bold shrink-0">3.</span>
            Cek ranking-mu di leaderboard web
          </li>
        </ol>
        <p className="mt-3 text-[11px] font-mono text-gray-500">
          Semua command bisa pakai awalan <span className="text-gray-300">/</span>{" "}
          atau <span className="text-gray-300">#</span> — fungsinya identik.
          Bot hanya merespon di grup yang sudah dikonfigurasi.
        </p>
      </div>

      {WHATSAPP_CATEGORIES.map((category, i) => (
        <CategorySection key={category.id} category={category} index={i} />
      ))}
    </div>
  );
}

function WebsiteGuide() {
  return (
    <div className="space-y-10">
      <section aria-labelledby="website-steps-heading">
        <header className="mb-4">
          <p className="font-orbitron text-[10px] uppercase tracking-[0.25em] text-system-blue/80">
            01 / Cara Pakai Website
          </p>
          <h3
            id="website-steps-heading"
            className="mt-1.5 text-base sm:text-lg font-bold font-orbitron text-white flex items-center gap-2"
          >
            <Globe size={18} className="text-system-blue shrink-0" />
            Dari Penonton Jadi Hunter
          </h3>
          <p className="mt-1 text-xs text-gray-500 font-mono leading-relaxed max-w-2xl">
            {WEBSITE_URL} — semua fitur di bawah berbagi data yang sama dengan
            bot WhatsApp.
          </p>
        </header>
        <ol className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {WEBSITE_STEPS.map((step, i) => {
            const Icon = step.icon;
            return (
              <li
                key={step.title}
                className="glass rounded-2xl p-4 sm:p-5 border border-gray-800 relative overflow-hidden"
              >
                <span
                  aria-hidden="true"
                  className="absolute -top-2 -right-1 font-orbitron text-5xl font-black text-gray-800/50 select-none"
                >
                  {i + 1}
                </span>
                <Icon size={20} className="text-system-volt" />
                <h4 className="mt-3 text-sm font-bold font-orbitron text-white">
                  {step.title}
                </h4>
                <p className="mt-1.5 text-xs text-gray-400 leading-relaxed">
                  {step.description}
                </p>
              </li>
            );
          })}
        </ol>
      </section>

      <section aria-labelledby="website-note-heading">
        <div className="glass rounded-2xl p-4 sm:p-5 border border-system-blue/25 flex items-start gap-3">
          <CalendarClock
            size={18}
            className="text-system-blue shrink-0 mt-0.5"
          />
          <div>
            <h4
              id="website-note-heading"
              className="text-sm font-bold font-orbitron text-white"
            >
              Satu Akun, Dua Pintu
            </h4>
            <p className="mt-1 text-xs text-gray-400 leading-relaxed font-mono">
              Laporan dari web dan WhatsApp masuk ke streak, XP, dan leaderboard
              yang sama. Kuota 3x/hari juga dibagi — lapor lewat mana pun,
              progresnya tetap satu.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Page                                                                        */
/* -------------------------------------------------------------------------- */

type Platform = "whatsapp" | "website";

const PLATFORMS: { id: Platform; label: string; icon: LucideIcon }[] = [
  { id: "whatsapp", label: "Bot WhatsApp", icon: MessageCircle },
  { id: "website", label: "Website", icon: Globe },
];

interface CommandsPageProps {
  /** Navigate back to the leaderboard dashboard. */
  onBack: () => void;
}

/**
 * Public "Panduan & Commands" page — reachable from the leaderboard so people
 * who have never used the bot can learn every WhatsApp command and website
 * feature before joining.
 */
export function CommandsPage({ onBack }: CommandsPageProps) {
  const [platform, setPlatform] = useState<Platform>("whatsapp");

  return (
    <div className="max-w-4xl mx-auto">
      {/* Hero */}
      <div className="mb-8">
        <button
          type="button"
          onClick={onBack}
          className="btn-press flex items-center gap-1.5 min-h-[44px] px-3 -ml-3 rounded-xl text-xs font-mono text-gray-400 hover:text-white transition-colors"
        >
          <ChevronLeft size={14} />
          Kembali ke Leaderboard
        </button>
        <div className="mt-2 flex items-start gap-3">
          <div className="shrink-0 w-11 h-11 rounded-2xl bg-system-green/10 border border-system-green/30 flex items-center justify-center shadow-neon-green">
            <Terminal size={20} className="text-system-green" />
          </div>
          <div className="min-w-0">
            <h2 className="text-xl sm:text-2xl font-black font-orbitron tracking-wider text-white uppercase">
              Panduan &amp; Commands
            </h2>
            <p className="mt-1 text-xs sm:text-sm text-gray-500 font-mono leading-relaxed">
              Semua cara memakai Lapor Bot — di WhatsApp dan di website. Klik
              command mana pun untuk menyalinnya.
            </p>
          </div>
        </div>
      </div>

      {/* Platform toggle — mirrors the ProgramTabBar pill pattern */}
      <div
        role="tablist"
        aria-label="Pilih platform panduan"
        className="mb-8 flex flex-nowrap overflow-x-auto scrollbar-none bg-gray-950/80 p-1.5 rounded-xl border border-gray-800/60 w-full sm:w-auto sm:max-w-fit"
      >
        {PLATFORMS.map((p) => {
          const Icon = p.icon;
          const active = platform === p.id;
          return (
            <button
              key={p.id}
              role="tab"
              id={`commands-tab-${p.id}`}
              aria-selected={active}
              aria-controls={`commands-panel-${p.id}`}
              onClick={() => setPlatform(p.id)}
              className={`flex shrink-0 whitespace-nowrap items-center gap-2 px-4 py-2 min-h-[40px] rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
                active
                  ? "bg-gradient-to-r from-system-green to-system-volt text-gray-950 shadow-neon-volt"
                  : "text-gray-400 hover:text-white hover:bg-gray-800/60"
              }`}
            >
              <Icon size={14} />
              {p.label}
            </button>
          );
        })}
      </div>

      <div
        role="tabpanel"
        id={`commands-panel-${platform}`}
        aria-labelledby={`commands-tab-${platform}`}
      >
        {platform === "whatsapp" ? <WhatsAppGuide /> : <WebsiteGuide />}
      </div>

      {/* Footer CTA */}
      <div className="mt-12 glass glass-glow-gold rounded-2xl p-5 sm:p-6 border border-gray-800 text-center">
        <Footprints className="mx-auto text-system-gold mb-3" size={24} />
        <p className="font-orbitron text-sm font-bold text-white uppercase tracking-wider">
          Siap Mulai?
        </p>
        <p className="mt-2 text-xs text-gray-400 font-mono leading-relaxed max-w-md mx-auto">
          Kirim <span className="text-system-green font-bold">/lapor</span>{" "}
          pertamamu di grup WhatsApp, lalu kembali ke sini untuk melihat
          namamu naik di klasemen.
        </p>
        <button
          type="button"
          onClick={onBack}
          className="btn-press mt-4 inline-flex items-center gap-2 px-5 py-2.5 min-h-[44px] rounded-xl bg-gradient-to-r from-system-green to-system-volt border border-system-volt/40 font-orbitron text-xs font-bold text-gray-950 uppercase tracking-wider shadow-neon-volt hover:brightness-110 transition-all"
        >
          <Trophy size={14} />
          Lihat Leaderboard
        </button>
      </div>
    </div>
  );
}
