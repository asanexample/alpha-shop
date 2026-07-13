import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AuthProvider } from "./context/AuthContext";
import { CartProvider } from "./context/CartContext";
import { Cart } from "./pages/Cart";
import { Category } from "./pages/Category";
import { Checkout } from "./pages/Checkout";
import { Home } from "./pages/Home";
import { Info } from "./pages/Info";
import { Login } from "./pages/Login";
import { NotFound } from "./pages/NotFound";
import { Order } from "./pages/Order";
import { Orders } from "./pages/Orders";
import { ProductDetail } from "./pages/Product";
import { Search } from "./pages/Search";
import { Signup } from "./pages/Signup";

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
      <AuthProvider>
        <CartProvider>
          <BrowserRouter>
            <Routes>
              <Route element={<Layout />}>
                <Route index element={<Home />} />
                <Route path="/c/:slug" element={<Category />} />
                <Route path="/p/:slug" element={<ProductDetail />} />
                <Route path="/cart" element={<Cart />} />
                <Route path="/cart/checkout" element={<Checkout />} />
                <Route path="/order/:id" element={<Order />} />
                <Route path="/account/orders" element={<Orders />} />
                <Route path="/search" element={<Search />} />
                <Route path="/login" element={<Login />} />
                <Route path="/signup" element={<Signup />} />
                <Route path="/service" element={<Info page="service" />} />
                <Route path="/about" element={<Info page="about" />} />
                <Route path="/community" element={<Info page="community" />} />
                <Route path="*" element={<NotFound />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </CartProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}
