import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { Navigate, useLocation } from "react-router";
import { api, ApiError, setCsrfToken, type Me } from "./api";

type AuthState = {
	me: Me | null;
	loading: boolean;
	refresh: () => Promise<Me | null>;
	login: (email: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [me, setMe] = useState<Me | null>(null);
	const [loading, setLoading] = useState(true);

	const refresh = useCallback(async () => {
		try {
			const m = await api.me();
			setCsrfToken(m.csrfToken);
			setMe(m);
			return m;
		} catch (err) {
			if (!(err instanceof ApiError && err.status === 401)) console.error(err);
			setCsrfToken("");
			setMe(null);
			return null;
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		refresh();
	}, [refresh]);

	const login = useCallback(
		async (email: string, password: string) => {
			await api.login(email, password);
			await refresh();
		},
		[refresh],
	);

	const logout = useCallback(async () => {
		try {
			await api.logout();
		} finally {
			setCsrfToken("");
			setMe(null);
		}
	}, []);

	return <AuthContext.Provider value={{ me, loading, refresh, login, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
	return ctx;
}

/** Logged-in session data; only call inside <RequireAuth>. */
export function useMe(): Me {
	const { me } = useAuth();
	if (!me) throw new Error("useMe called without a session");
	return me;
}

export function RequireAuth({ children }: { children: ReactNode }) {
	const { me, loading } = useAuth();
	const location = useLocation();
	if (loading) return null;
	if (!me) return <Navigate to="/login" replace state={{ from: location.pathname }} />;
	return children;
}

/** For /login and /register: bounce signed-in users to the dashboard. */
export function GuestOnly({ children }: { children: ReactNode }) {
	const { me, loading } = useAuth();
	if (loading) return null;
	if (me) return <Navigate to="/dashboard" replace />;
	return children;
}
