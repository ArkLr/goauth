import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router";
import { GuestOnly, RequireAuth } from "./auth";
import Dashboard from "./pages/Dashboard";
import EditPost from "./pages/EditPost";
import Login from "./pages/Login";
import NotFound from "./pages/NotFound";
import Posts from "./pages/Posts";
import Register from "./pages/Register";

const titles: Record<string, string> = {
	"/login": "Sign in",
	"/register": "Create account",
	"/dashboard": "Dashboard",
	"/posts": "Posts",
};

export default function App() {
	const { pathname } = useLocation();
	useEffect(() => {
		const t = titles[pathname] ?? (pathname.endsWith("/edit") ? "Edit post" : null);
		document.title = t ? `${t} · GoAuth` : "GoAuth";
	}, [pathname]);

	return (
		<div className="flex min-h-screen items-center justify-center px-4 py-6">
			<Routes>
				<Route path="/" element={<Navigate to="/dashboard" replace />} />
				<Route path="/protected" element={<Navigate to="/dashboard" replace />} />
				<Route path="/login" element={<GuestOnly><Login /></GuestOnly>} />
				<Route path="/register" element={<GuestOnly><Register /></GuestOnly>} />
				<Route path="/dashboard" element={<RequireAuth><Dashboard /></RequireAuth>} />
				<Route path="/posts" element={<RequireAuth><Posts /></RequireAuth>} />
				<Route path="/posts/:id/edit" element={<RequireAuth><EditPost /></RequireAuth>} />
				<Route path="*" element={<NotFound />} />
			</Routes>
		</div>
	);
}
