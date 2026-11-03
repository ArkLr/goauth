// Thin fetch wrapper for the Go JSON API.
//
// The session lives in an HttpOnly cookie the browser sends automatically.
// State-changing requests also need the CSRF token from GET /api/me, which
// AuthProvider stores here via setCsrfToken().

export type User = {
	id: number;
	email: string;
	createdAt: string;
};

export type Post = {
	id: number;
	userId: number;
	headline: string;
	body: string;
	createdAt: string;
	updatedAt: string;
	authorEmail?: string;
};

export type Me = {
	user: User;
	csrfToken: string;
	expiresAt: string;
};

export type Dashboard = {
	postCount: number;
	recentPosts: Post[];
};

export class ApiError extends Error {
	status: number;
	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

let csrfToken = "";

export function setCsrfToken(token: string) {
	csrfToken = token;
}

async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = { Accept: "application/json" };
	if (body !== undefined) headers["Content-Type"] = "application/json";
	if (method !== "GET" && csrfToken) headers["X-CSRF-Token"] = csrfToken;

	const res = await fetch(url, {
		method,
		headers,
		credentials: "same-origin",
		body: body === undefined ? undefined : JSON.stringify(body),
	});

	if (!res.ok) {
		let message = "something went wrong, try again";
		try {
			const data = await res.json();
			if (data?.error) message = data.error;
		} catch {
			// non-JSON error body; keep the generic message
		}
		throw new ApiError(res.status, message);
	}

	if (res.status === 204 || res.headers.get("Content-Length") === "0") {
		return undefined as T;
	}
	const text = await res.text();
	return (text ? JSON.parse(text) : undefined) as T;
}

export const api = {
	me: () => request<Me>("GET", "/api/me"),
	login: (email: string, password: string) => request<void>("POST", "/api/login", { email, password }),
	register: (email: string, password: string, confirmPassword: string) =>
		request<void>("POST", "/api/register", { email, password, confirmPassword }),
	logout: () => request<void>("POST", "/api/logout"),

	dashboard: () => request<Dashboard>("GET", "/api/dashboard"),

	listPosts: (mine: boolean) =>
		request<{ posts: Post[] }>("GET", mine ? "/api/posts?filter=mine" : "/api/posts").then((r) => r.posts),
	getPost: (id: number) => request<Post>("GET", `/api/posts/${id}`),
	createPost: (headline: string, body: string) => request<Post>("POST", "/api/posts", { headline, body }),
	updatePost: (id: number, headline: string, body: string) =>
		request<Post>("PUT", `/api/posts/${id}`, { headline, body }),
	deletePost: (id: number) => request<void>("DELETE", `/api/posts/${id}`),
};

export function errorMessage(err: unknown): string {
	return err instanceof Error ? err.message : "something went wrong, try again";
}

// "Jan 2, 2006 · 15:04" (or just "Jan 2, 2006"), matching the old templates.
export function formatDate(iso: string, withTime = true): string {
	const d = new Date(iso);
	const date = d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
	if (!withTime) return date;
	const time = d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false });
	return `${date} · ${time}`;
}

export function isEdited(p: Post): boolean {
	return p.updatedAt !== p.createdAt;
}
