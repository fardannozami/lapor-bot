import { useEffect, useRef, useState } from "react";
import { useRepositories } from "@lapor-bot/shared";
import type { StravaStatus } from "@lapor-bot/shared";

interface StravaConnectPanelProps {
  onMessage?: (message: string, tone?: "success" | "error") => void;
}

function formatExpiry(expiresAt?: string) {
  if (!expiresAt) return "";
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(expiresAt));
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function StravaConnectPanel({ onMessage }: StravaConnectPanelProps) {
  const { reports: repo } = useRepositories();
  const [status, setStatus] = useState<StravaStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [disconnecting, setDisconnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const popupRef = useRef<Window | null>(null);
  const allowedCallbackOriginRef = useRef<string | null>(null);

  const refreshStatus = async () => {
    const nextStatus = await repo.getStravaStatus();
    setStatus(nextStatus);
    return nextStatus;
  };

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    refreshStatus()
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(errorMessage(err, "Gagal memuat status Strava"));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const handleMessage = async (event: MessageEvent) => {
      const allowedCallbackOrigin = allowedCallbackOriginRef.current;
      if (
        event.origin !== window.location.origin &&
        event.origin !== allowedCallbackOrigin
      ) {
        return;
      }
      if (event.data?.type !== "lapor-bot:strava-connected") return;

      setConnecting(false);
      try {
        const nextStatus = await refreshStatus();
        if (nextStatus.linked) {
          onMessage?.("Strava berhasil terhubung");
        }
      } catch (err: unknown) {
        setError(errorMessage(err, "Gagal memuat ulang status Strava"));
      }
    };

    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [onMessage]);

  const connect = async () => {
    setConnecting(true);
    setError(null);
    try {
      const { auth_url: authUrl } = await repo.createStravaLink();
      const redirectUri = new URL(authUrl).searchParams.get("redirect_uri");
      allowedCallbackOriginRef.current = redirectUri
        ? new URL(redirectUri).origin
        : null;
      const popup = window.open(
        authUrl,
        "lapor-bot-strava",
        "popup,width=520,height=720",
      );

      if (!popup) {
        window.location.assign(authUrl);
        return;
      }

      popupRef.current = popup;
      const startedAt = Date.now();
      const poll = window.setInterval(async () => {
        if (popup.closed || Date.now() - startedAt > 120_000) {
          window.clearInterval(poll);
          setConnecting(false);
          try {
            const nextStatus = await refreshStatus();
            if (nextStatus.linked) onMessage?.("Strava berhasil terhubung");
          } catch {
            return;
          }
        }
      }, 1_000);
    } catch (err: unknown) {
      setError(errorMessage(err, "Gagal membuat link Strava"));
      setConnecting(false);
    }
  };

  const disconnect = async () => {
    setDisconnecting(true);
    setError(null);
    try {
      const result = await repo.disconnectStrava();
      await refreshStatus();
      onMessage?.(result.message || "Strava berhasil diputus");
    } catch (err: unknown) {
      const message = errorMessage(err, "Gagal memutus Strava");
      setError(message);
      onMessage?.(message, "error");
    } finally {
      setDisconnecting(false);
    }
  };

  const linked = status?.linked ?? false;

  return (
    <section className="glass rounded-3xl p-4 sm:p-6 border border-orange-500/20">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
            <span aria-hidden="true">🟧</span>
            Strava Sync
          </h3>
          <p className="text-xs text-gray-500 font-mono mt-1 leading-relaxed">
            Hubungkan Strava agar aktivitas baru bisa otomatis masuk ke Lapor
            Bot.
          </p>
        </div>
        <span
          className={`shrink-0 rounded-full border px-2.5 py-1 text-[10px] font-bold font-mono uppercase ${
            linked
              ? "border-system-green/40 bg-system-green/10 text-system-green"
              : "border-gray-700 bg-gray-950/60 text-gray-500"
          }`}
        >
          {loading ? "Checking" : linked ? "Connected" : "Offline"}
        </span>
      </div>

      <div className="mt-4 rounded-2xl border border-gray-800 bg-gray-950/50 p-4">
        {loading ? (
          <p className="text-xs text-gray-500 font-mono">
            Memuat status Strava...
          </p>
        ) : linked ? (
          <div className="space-y-1">
            <p className="text-sm font-bold text-white font-mono">
              Strava aktif{status?.name ? ` untuk ${status.name}` : ""}
            </p>
            <p className="text-xs text-gray-500 font-mono">
              Athlete ID: {status?.athlete_id ?? "—"}
            </p>
            {status?.expires_at && (
              <p className="text-[10px] text-gray-600 font-mono uppercase">
                Token refresh sebelum {formatExpiry(status.expires_at)}
              </p>
            )}
          </div>
        ) : (
          <p className="text-sm text-gray-400 font-mono">
            Belum terhubung. Klik tombol di bawah untuk login Strava dan memberi
            akses baca aktivitas.
          </p>
        )}
      </div>

      {error && (
        <p className="mt-3 text-[10px] text-system-red font-mono">{error}</p>
      )}

      <div className="mt-4 flex flex-col sm:flex-row gap-2">
        {linked ? (
          <button
            type="button"
            onClick={disconnect}
            disabled={disconnecting}
            className="px-4 py-2 rounded-xl bg-system-red/10 hover:bg-system-red/20 border border-system-red/20 text-system-red font-mono text-xs font-bold transition-colors disabled:opacity-50"
          >
            {disconnecting ? "Memutus..." : "Disconnect Strava"}
          </button>
        ) : (
          <button
            type="button"
            onClick={connect}
            disabled={connecting}
            className="px-4 py-2 rounded-xl bg-orange-500/15 hover:bg-orange-500/25 border border-orange-500/30 text-orange-300 font-mono text-xs font-bold transition-colors disabled:opacity-50"
          >
            {connecting ? "Menunggu Strava..." : "Connect Strava"}
          </button>
        )}
        <button
          type="button"
          onClick={async () => {
            setLoading(true);
            setError(null);
            try {
              await refreshStatus();
            } catch (err: unknown) {
              setError(errorMessage(err, "Gagal refresh status Strava"));
            } finally {
              setLoading(false);
            }
          }}
          className="px-4 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 text-gray-400 hover:text-white font-mono text-xs transition-colors"
        >
          Refresh Status
        </button>
      </div>
    </section>
  );
}
