import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { api, errorMessage } from "../api";
import { Alert, Button, Card, CardBody, CardFooter, Field, Input, PasswordInput, Title } from "../components/ui";

export default function Register() {
	const navigate = useNavigate();

	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const mismatch = confirm !== "" && confirm !== password;

	async function onSubmit(e: FormEvent) {
		e.preventDefault();
		setError("");
		if (mismatch) {
			setError("passwords do not match");
			return;
		}
		setBusy(true);
		try {
			await api.register(email, password, confirm);
			navigate("/login?registered=1");
		} catch (err) {
			setError(errorMessage(err));
		} finally {
			setBusy(false);
		}
	}

	return (
		<Card>
			<CardBody>
				<Title subtitle="Enter an email and choose a password.">Create your account</Title>

				<Alert>{error}</Alert>

				<form onSubmit={onSubmit}>
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

					<Field label="Password" htmlFor="password" hint="Between 8 and 72 characters.">
						<PasswordInput
							id="password"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
							placeholder="At least 8 characters"
							autoComplete="new-password"
							minLength={8}
							maxLength={72}
							required
						/>
					</Field>

					<Field
						label="Confirm password"
						htmlFor="confirm_password"
						hint={mismatch ? "Passwords do not match." : "Type the same password again."}
					>
						<PasswordInput
							id="confirm_password"
							value={confirm}
							onChange={(e) => setConfirm(e.target.value)}
							placeholder="Repeat your password"
							autoComplete="new-password"
							minLength={8}
							maxLength={72}
							required
							aria-invalid={mismatch}
						/>
					</Field>

					<Button type="submit" block disabled={busy}>
						{busy ? "Creating account…" : "Create account"}
					</Button>
				</form>
			</CardBody>

			<CardFooter>
				Have an account? <Link to="/login">Sign in</Link>
			</CardFooter>
		</Card>
	);
}
