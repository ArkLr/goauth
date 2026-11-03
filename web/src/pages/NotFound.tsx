import { Link } from "react-router";
import { Card, CardBody, Title } from "../components/ui";

export default function NotFound() {
	return (
		<Card>
			<CardBody>
				<Title subtitle="That page doesn't exist.">Not found</Title>
				<Link to="/" className="text-sm font-semibold text-fg hover:underline">
					Go home
				</Link>
			</CardBody>
		</Card>
	);
}
