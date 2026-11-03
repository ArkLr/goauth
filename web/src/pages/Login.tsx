import { useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
import { errorMessage } from "../api";
import { useAuth } from "../auth";
import { Alert, Button, Card, CardBody, CardFooter, Field, Input, PasswordInput, Title } from "../components/ui";

export default function Login() {
	const { login } = useAuth();
	const navigate = useNavigate();
	const location = useLocation();
	const [params] = useSearchParams();

	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	async function onSubmit(e: FormEvent) {
		e.preventDefault();
		setError("");
		if (!email.trim() || !password) {
			setError("email and password are required");
			return;
		}
		setBusy(true);
		try {
			await login(email, password);
			const from = (location.state as { from?: string } | null)?.from;
			navigate(from ?? "/dashboard", { replace: true });
		} catch (err) {
			setError(errorMessage(err));
			setPassword("");
		} finally {
			setBusy(false);
		}
	}

	return (
		<Card>
			<CardBody>
				<Title subtitle="Enter your email and password to continue.">Sign in to GoAuth</Title>

				<Alert>{error}</Alert>
				{!error && params.get("registered") === "1" && (
					<Alert kind="success">Account created. Please log in.</Alert>
				)}

				<form onSubmit={onSubmit} noValidate>
					<Field label="Email" htmlFor="email">
						<Input
							type="email"
							id="email"
							value={email}
							onChange={(e) => setEmail(e.target.value)}
							placeholder="name@example.com"
							autoComplete="email"
							required
							autoFocus
						/>
					</Field>

					<Field label="Password" htmlFor="password">
						<PasswordInput
							id="password"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
							placeholder="Enter your password"
							autoComplete="current-password"
							required
						/>
					</Field>

					<Button type="submit" block disabled={busy}>
						{busy ? "Signing in…" : "Continue"}
					</Button>
				</form>
			</CardBody>

			<CardFooter>
				No account? <Link to="/register">Sign up</Link>
			</CardFooter>
		</Card>
	);
}
