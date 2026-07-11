import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { CartProvider } from "./context/CartContext";
import { Category } from "./pages/Category";
import { Home } from "./pages/Home";
import { Info } from "./pages/Info";
import { NotFound } from "./pages/NotFound";
import { ProductDetail } from "./pages/Product";
import { Search } from "./pages/Search";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      gcTime: 5 * 60_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <CartProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route index element={<Home />} />
              <Route path="/c/:slug" element={<Category />} />
              <Route path="/p/:slug" element={<ProductDetail />} />
              <Route path="/search" element={<Search />} />
              <Route path="/service" element={<Info page="service" />} />
              <Route path="/about" element={<Info page="about" />} />
              <Route path="/community" element={<Info page="community" />} />
              <Route path="*" element={<NotFound />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </CartProvider>
    </QueryClientProvider>
  );
}
