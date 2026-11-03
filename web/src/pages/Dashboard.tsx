import { useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate } from "react-router";
import {
	api,
	errorMessage,
	formatDate,
	isEdited,
	type Dashboard as DashboardData,
} from "../api";
import { useAuth, useMe } from "../auth";
import { ListIcon, LogOutIcon, PlusIcon, UserIcon } from "../components/icons";
import { Alert, Button, ButtonLink } from "../components/ui";

export default function Dashboard() {
	const me = useMe();
	const { logout } = useAuth();
	const navigate = useNavigate();

	const [data, setData] = useState<DashboardData | null>(null);
	const [error, setError] = useState("");

	useEffect(() => {
		api.dashboard().then(setData, (err) => setError(errorMessage(err)));
	}, []);

	async function onLogout() {
		try {
			await logout();
		} finally {
			navigate("/login", { replace: true });
		}
	}

	return (
		<div className="w-full max-w-full  px-8 py-8">
			<div className="mb-8 flex flex-col gap-4">
				<div className="flex flex-row items-center gap-3">
					<div className="flex size-12 shrink-0 items-center justify-center rounded-full border border-line bg-surface">
						<UserIcon size={24} />
					</div>
					<div className="min-w-0">
						<h1 className="text-2xl font-medium tracking-[-0.02em]">
							Username
						</h1>
						<p className="truncate text-sm text-muted">{me.user.email}</p>
					</div>
				</div>

				<div className="flex flex-wrap gap-2">
					<Stat label="Member since">
						{formatDate(me.user.createdAt, false)}
					</Stat>
					<Stat label="Status">
						<span className="inline-flex items-center gap-1.5 rounded-full border border-green-500/28 bg-green-500/8 px-2.5 py-0.5 text-xs font-medium text-success-text">
							<span className="size-1.5 rounded-full bg-success-text" />
							Active session
						</span>
					</Stat>
					<Stat label="Session expires">{formatDate(me.expiresAt)}</Stat>
					<Stat label="Number posts">{data ? data.postCount : "–"}</Stat>
				</div>
			</div>

			<Alert>{error}</Alert>

			{/* Quick actions */}
			<div className="mb-8 flex w-full flex-col gap-4">
				<SectionTitle>Quick actions</SectionTitle>
				<Action to="/posts#compose" icon={<PlusIcon />} title="New post" />
				<Action
					to="/posts"
					icon={<ListIcon />}
					title="All posts"
					desc="See what's new"
				/>
				<Action to="/posts?filter=mine" icon={<UserIcon />} title="My posts" />
			</div>

			{/* Recent posts */}
			<div className="flex items-baseline justify-between">
				<SectionTitle>Your recent posts</SectionTitle>
				{data && data.recentPosts.length > 0 && (
					<Link
						to="/posts?filter=mine"
						className="text-sm text-muted hover:text-fg"
					>
						View all
					</Link>
				)}
			</div>
			<ul className="mb-8 divide-y divide-card-line overflow-hidden rounded-field border border-line">
				{data?.recentPosts.map((p) => (
					<li
						key={p.id}
						className="flex items-center justify-between gap-3 px-4 py-3"
					>
						<div className="min-w-0">
							<div className="truncate font-medium">{p.headline}</div>
							<div className="text-xs text-muted">
								{formatDate(p.createdAt)}
								{isEdited(p) && " · edited"}
							</div>
						</div>
						<ButtonLink to={`/posts/${p.id}/edit`} small>
							Edit
						</ButtonLink>
					</li>
				))}
				{data && data.recentPosts.length === 0 && (
					<li className="px-4 py-5 text-sm text-muted">
						Nothing yet.{" "}
						<Link
							to="/posts#compose"
							className="font-semibold text-fg hover:underline"
						>
							Write your first post
						</Link>
					</li>
				)}
				{!data && !error && (
					<li className="px-4 py-5 text-sm text-muted">Loading…</li>
				)}
			</ul>

			<Button variant="outline" onClick={onLogout}>
				<LogOutIcon />
				Sign out
			</Button>
		</div>
	);
}

function Stat({ label, children }: { label: string; children: ReactNode }) {
	return (
		<div className="flex flex-col items-center justify-center rounded-field bg-surface px-4 py-3.5">
			<div className="mb-1 text-xs text-muted">{label}</div>
			<div className="flex min-h-6 items-center font-medium">{children}</div>
		</div>
	);
}

function SectionTitle({ children }: { children: ReactNode }) {
	return (
		<h2 className="mb-3 text-sm font-semibold tracking-wide text-muted uppercase">
			{children}
		</h2>
	);
}

function Action({
	to,
	icon,
	title,
	desc,
}: {
	to: string;
	icon: ReactNode;
	title: string;
	desc?: string;
}) {
	return (
		<Link
			to={to}
			className="flex flex-col gap-1 rounded-field border border-line bg-surface p-4 text-fg no-underline transition hover:border-line-hover"
		>
			<span className="mb-1 flex size-8 items-center justify-center rounded-lg bg-card">
				{icon}
			</span>
			<span className="font-medium">{title}</span>
			{desc && <span className="text-xs text-muted">{desc}</span>}
		</Link>
	);
}
