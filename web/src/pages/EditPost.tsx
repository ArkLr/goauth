import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { api, ApiError, errorMessage } from "../api";
import { Alert, BackLink, Button, Card, CardBody, CardFooter, Field, Input, Textarea, Title } from "../components/ui";
import NotFound from "./NotFound";

export default function EditPost() {
	const { id: idParam } = useParams();
	const id = Number(idParam);
	const navigate = useNavigate();

	const [loaded, setLoaded] = useState(false);
	const [notFound, setNotFound] = useState(false);
	const [headline, setHeadline] = useState("");
	const [body, setBody] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	useEffect(() => {
		if (!Number.isInteger(id) || id <= 0) {
			setNotFound(true);
			return;
		}
		api.getPost(id).then(
			(p) => {
				setHeadline(p.headline);
				setBody(p.body);
				setLoaded(true);
			},
			(err) => {
				if (err instanceof ApiError && err.status === 404) setNotFound(true);
				else setError(errorMessage(err));
			},
		);
	}, [id]);

	async function onSubmit(e: FormEvent) {
		e.preventDefault();
		setError("");
		setBusy(true);
		try {
			await api.updatePost(id, headline, body);
			navigate("/posts");
		} catch (err) {
			setError(errorMessage(err));
		} finally {
			setBusy(false);
		}
	}

	if (notFound) return <NotFound />;

	return (
		<Card wide>
			<CardBody>
				<BackLink to="/posts">Back to posts</BackLink>
				<Title subtitle="Update the headline or body, then save your changes.">Edit post</Title>

				<Alert>{error}</Alert>

				{loaded ? (
					<form onSubmit={onSubmit}>
						<Field label="Headline" htmlFor="headline">
							<Input
								type="text"
								id="headline"
								value={headline}
								onChange={(e) => setHeadline(e.target.value)}
								maxLength={200}
								required
								autoFocus
							/>
						</Field>
						<Field label="Body" htmlFor="body">
							<Textarea id="body" value={body} onChange={(e) => setBody(e.target.value)} rows={8} required />
						</Field>
						<Button type="submit" block disabled={busy}>
							{busy ? "Saving…" : "Save changes"}
						</Button>
					</form>
				) : (
					!error && <p className="text-sm text-muted">Loading…</p>
				)}
			</CardBody>

			<CardFooter>
				<Link to="/posts">Cancel</Link>
			</CardFooter>
		</Card>
	);
}
