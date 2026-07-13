import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { useAuth } from "../context/AuthContext";
import { ApiError } from "../lib/api";
import styles from "./Login.module.css";

export function Login() {
  const { login, isLoggingIn } = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const next = params.get("next") || "/";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await login({ email: email.trim(), password });
      navigate(next, { replace: true });
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Something went wrong signing you in. Please try again.",
      );
    }
  }

  return (
    <div className={styles.wrap}>
      <Breadcrumb items={[{ label: "Home", to: "/" }, { label: "Sign in" }]} />

      <div className={styles.card}>
        <h1 className={styles.title}>Sign in</h1>
        <p className={styles.lede}>Sign in to check out, and to see your order history.</p>

        <form className={styles.form} onSubmit={onSubmit} noValidate>
          <label className={styles.field}>
            <span className={styles.label}>Email</span>
            <input
              className={styles.input}
              type="email"
              autoComplete="email"
              required
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoggingIn}
            />
          </label>
          <label className={styles.field}>
            <span className={styles.label}>Password</span>
            <input
              className={styles.input}
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoggingIn}
            />
          </label>

          {error ? (
            <div className={styles.error} role="alert">
              {error}
            </div>
          ) : null}

          <button type="submit" className="btn btn--lg" disabled={isLoggingIn}>
            {isLoggingIn ? "Signing in…" : "Sign in"}
          </button>
        </form>

        <p className={styles.switch}>
          New to Alpha Bikes?{" "}
          <Link to={`/signup${next !== "/" ? `?next=${encodeURIComponent(next)}` : ""}`}>
            Create an account
          </Link>
        </p>
      </div>
    </div>
  );
}
