import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { Footer } from "./Footer";
import { Header } from "./Header";
import { Toaster } from "./Toaster";

// Reset scroll on navigation (SPA doesn't do this by default).
function ScrollToTop() {
  const { pathname, search } = useLocation();
  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: "instant" as ScrollBehavior });
  }, [pathname, search]);
  return null;
}

export function Layout() {
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <ScrollToTop />
      <Header />
      <main id="main" className="page">
        <Outlet />
      </main>
      <Footer />
      <Toaster />
    </>
  );
}
