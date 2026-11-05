import {
	useCallback,
	useEffect,
	useRef,
	useState,
	type FormEvent,
} from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
import { api, errorMessage, formatDate, isEdited, type Post } from "../api";
import { useMe } from "../auth";
import {
	Dialog,
	DialogContent,
	DialogTrigger,
} from "@/components/arc/dialog/dialog";
import { PencilIcon, PlusIcon, TrashIcon } from "../components/icons";
import {
	Alert,
	BackLink,
	Button,
	ButtonLink,
	Card,
	CardBody,
	CardSection,
	Field,
	Input,
	Textarea,
	Title,
} from "../components/ui";

export default function Posts() {
	const me = useMe();
	const location = useLocation();
	const navigate = useNavigate();
	const [params] = useSearchParams();
	const mine = params.get("filter") === "mine";

	const [composeOpen, setComposeOpen] = useState(false);
	const [posts, setPosts] = useState<Post[] | null>(null);
	const [listError, setListError] = useState("");

	const [headline, setHeadline] = useState("");
	const [body, setBody] = useState("");
	const [formError, setFormError] = useState("");
	const [busy, setBusy] = useState(false);
	const headlineRef = useRef<HTMLInputElement>(null);

	const load = useCallback(() => {
		setListError("");
		return api
			.listPosts(mine)
			.then(setPosts, (err) => setListError(errorMessage(err)));
	}, [mine]);

	useEffect(() => {
		setPosts(null);
		load();
	}, [load]);

	// Links to /posts#compose (from the dashboard) open the dialog straight away.
	useEffect(() => {
		if (location.hash === "#compose") setComposeOpen(true);
	}, [location.hash]);

	function onComposeChange(open: boolean) {
		setComposeOpen(open);
		if (open) setFormError("");
		// Drop the hash so a page reload doesn't reopen the dialog.
		else if (location.hash)
			navigate(location.pathname + location.search, { replace: true });
	}

	async function onCreate(e: FormEvent) {
		e.preventDefault();
		setFormError("");
		setBusy(true);
		try {
			await api.createPost(headline, body);
			setHeadline("");
			setBody("");
			onComposeChange(false);
			await load();
		} catch (err) {
			setFormError(errorMessage(err));
		} finally {
			setBusy(false);
		}
	}

	async function onDelete(id: number) {
		if (!confirm("Delete this post?")) return;
		try {
			await api.deletePost(id);
			setPosts((list) => list?.filter((p) => p.id !== id) ?? null);
		} catch (err) {
			setListError(errorMessage(err));
		}
	}

	const tab = (active: boolean) =>
		`rounded-lg px-3 py-1 text-sm no-underline transition ${active ? "bg-surface text-fg" : "text-muted hover:text-fg"}`;

	return (
		<Card wide>
			<CardBody>
				<BackLink to="/dashboard">Dashboard</BackLink>
				<Title subtitle="Share something with everyone signed in.">Posts</Title>

				<Dialog open={composeOpen} onOpenChange={onComposeChange}>
					<DialogTrigger asChild>
						<Button block>
							<PlusIcon />
							New post
						</Button>
					</DialogTrigger>
					<DialogContent
						className="arc-dialog"
						title="New post"
						description="Share something with everyone signed in."
						onOpenAutoFocus={(e) => {
							// Radix would otherwise focus the close button first.
							e.preventDefault();
							headlineRef.current?.focus();
						}}
					>
						<Alert>{formError}</Alert>

						<form id="compose" onSubmit={onCreate}>
							<Field label="Headline" htmlFor="headline">
								<Input
									ref={headlineRef}
									type="text"
									id="headline"
									value={headline}
									onChange={(e) => setHeadline(e.target.value)}
									placeholder="What's it about?"
									maxLength={200}
									required
								/>
							</Field>
							<Field label="Body" htmlFor="body">
								<Textarea
									id="body"
									value={body}
									onChange={(e) => setBody(e.target.value)}
									placeholder="Write your post..."
									rows={4}
									required
								/>
							</Field>
							<Button type="submit" block disabled={busy}>
								{busy ? "Publishing…" : "Publish post"}
							</Button>
						</form>
					</DialogContent>
				</Dialog>
			</CardBody>

			<CardSection>
				<div className="mb-4 flex items-center justify-between">
					<span className="text-sm text-muted">
						{posts
							? `${posts.length} ${posts.length === 1 ? "post" : "posts"}`
							: " "}
					</span>
					<nav
						aria-label="Filter posts"
						className="flex gap-1 rounded-field border border-line p-1"
					>
						<Link
							to="/posts"
							className={tab(!mine)}
							aria-current={!mine ? "page" : undefined}
						>
							All
						</Link>
						<Link
							to="/posts?filter=mine"
							className={tab(mine)}
							aria-current={mine ? "page" : undefined}
						>
							Mine
						</Link>
					</nav>
				</div>

				<Alert>{listError}</Alert>

				{posts === null && !listError && (
					<p className="py-6 text-center text-sm text-muted">Loading…</p>
				)}

				{posts?.length === 0 && (
					<p className="py-6 text-center text-sm text-muted">
						{mine ? "You haven't posted anything yet." : "No posts yet."}
					</p>
				)}

				<div className="divide-y divide-card-line">
					{posts?.map((p) => (
						<PostItem
							key={p.id}
							post={p}
							own={p.userId === me.user.id}
							onDelete={() => onDelete(p.id)}
						/>
					))}
				</div>
			</CardSection>
		</Card>
	);
}

function PostItem({
	post,
	own,
	onDelete,
}: {
	post: Post;
	own: boolean;
	onDelete: () => void;
}) {
	const author = post.authorEmail ?? "";
	return (
		<article className="py-5 first:pt-0 last:pb-0">
			<header className="mb-3 flex items-center gap-3">
				<div
					aria-hidden="true"
					className="flex size-9 shrink-0 items-center justify-center rounded-full border border-line bg-surface text-sm font-semibold uppercase"
				>
					{author.slice(0, 1)}
				</div>
				<div className="min-w-0">
					<div className="flex items-center gap-2 text-sm font-medium">
						<span className="truncate">{author}</span>
						{own && (
							<span className="rounded-full border border-line px-2 py-px text-[11px] font-medium text-muted">
								You
							</span>
						)}
					</div>
					<time dateTime={post.createdAt} className="text-xs text-muted">
						{formatDate(post.createdAt)}
						{isEdited(post) && " · edited"}
					</time>
				</div>
			</header>

			<h2 className="mb-1.5 text-lg font-semibold tracking-[-0.01em]">
				{post.headline}
			</h2>
			<p className="leading-relaxed break-words whitespace-pre-wrap text-fg/90">
				{post.body}
			</p>

			{own && (
				<div className="mt-4 flex gap-2">
					<ButtonLink to={`/posts/${post.id}/edit`} small>
						<PencilIcon />
						Edit
					</ButtonLink>
					<Button variant="danger" small onClick={onDelete}>
						<TrashIcon />
						Delete
					</Button>
				</div>
			)}
		</article>
	);
}
