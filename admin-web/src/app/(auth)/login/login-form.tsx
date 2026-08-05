"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

type LoginState = {
  email: string;
  password: string;
  error: string;
  isLoading: boolean;
  showPassword: boolean;
};

export function LoginForm() {
  const router = useRouter();
  const [state, setState] = useState<LoginState>({
    email: "",
    password: "",
    error: "",
    isLoading: false,
    showPassword: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const email = state.email.trim();
    if (!email || !state.password) {
      setState((current) => ({
        ...current,
        error: "Email dan password wajib diisi.",
      }));
      return;
    }

    setState((current) => ({ ...current, error: "", isLoading: true }));

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email,
          password: state.password,
        }),
      });

      const payload = (await response.json()) as {
        status?: string;
        message?: string;
      };

      if (!response.ok) {
        setState((current) => ({
          ...current,
          error: payload.message ?? "Login gagal. Coba lagi.",
          isLoading: false,
        }));
        return;
      }

      router.replace("/dashboard");
      router.refresh();
    } catch {
      setState((current) => ({
        ...current,
        error: "Layanan admin belum tersedia. Coba lagi nanti.",
        isLoading: false,
      }));
    }
  }

  return (
    <form className="mt-6 space-y-5" onSubmit={handleSubmit}>
      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-800" htmlFor="email">
          Email
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          value={state.email}
          onChange={(event) =>
            setState((current) => ({
              ...current,
              email: event.target.value,
            }))
          }
          className="h-11 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100"
          placeholder="admin@contoh.local"
        />
      </div>

      <div className="space-y-2">
        <label
          className="text-sm font-medium text-slate-800"
          htmlFor="password"
        >
          Password
        </label>
        <div className="flex h-11 overflow-hidden rounded-md border border-slate-300 focus-within:border-emerald-700 focus-within:ring-2 focus-within:ring-emerald-100">
          <input
            id="password"
            name="password"
            type={state.showPassword ? "text" : "password"}
            autoComplete="current-password"
            value={state.password}
            onChange={(event) =>
              setState((current) => ({
                ...current,
                password: event.target.value,
              }))
            }
            className="min-w-0 flex-1 px-3 text-sm outline-none"
            placeholder="Password admin"
          />
          <button
            type="button"
            className="w-20 border-l border-slate-200 text-sm font-medium text-slate-700 hover:bg-slate-50"
            onClick={() =>
              setState((current) => ({
                ...current,
                showPassword: !current.showPassword,
              }))
            }
          >
            {state.showPassword ? "Sembunyi" : "Lihat"}
          </button>
        </div>
      </div>

      {state.error ? (
        <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {state.error}
        </p>
      ) : null}

      <button
        type="submit"
        disabled={state.isLoading}
        className="h-11 w-full rounded-md bg-emerald-700 text-sm font-semibold text-white transition hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-400"
      >
        {state.isLoading ? "Memproses..." : "Masuk"}
      </button>
    </form>
  );
}
