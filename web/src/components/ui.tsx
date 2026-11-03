import { useState, type ButtonHTMLAttributes, type ComponentProps, type InputHTMLAttributes, type ReactNode, type TextareaHTMLAttributes } from "react";
import { Link, type LinkProps } from "react-router";
import { AlertIcon, CheckIcon, ChevronLeftIcon, EyeIcon, EyeOffIcon } from "./icons";

const cx = (...c: (string | false | null | undefined)[]) => c.filter(Boolean).join(" ");

/* ---------- Card ---------- */

export function Card({ wide, children }: { wide?: boolean; children: ReactNode }) {
	return (
		<main
			className={cx(
				"w-full animate-rise overflow-hidden rounded-card border border-card-line bg-card shadow-[0_24px_48px_rgba(0,0,0,0.25)]",
				wide ? "max-w-[560px]" : "max-w-[432px]",
			)}
		>
			{children}
		</main>
	);
}

export function CardBody({ children }: { children: ReactNode }) {
	return <div className="p-6 sm:p-8">{children}</div>;
}

export function CardSection({ children }: { children: ReactNode }) {
	return <div className="border-t border-card-line p-6 sm:px-8">{children}</div>;
}

export function CardFooter({ children }: { children: ReactNode }) {
	return (
		<div className="border-t border-card-line px-6 py-[18px] text-right text-[0.8rem] text-muted sm:px-8 [&_a]:font-semibold [&_a]:text-fg [&_a]:no-underline [&_a:hover]:underline">
			{children}
		</div>
	);
}

export function Title({ children, subtitle }: { children: ReactNode; subtitle?: ReactNode }) {
	return (
		<>
			<h1 className="mb-2 text-[2.1rem] leading-[1.15] font-medium tracking-[-0.03em]">{children}</h1>
			{subtitle && <p className="mb-7 text-muted">{subtitle}</p>}
		</>
	);
}

export function BackLink({ to, children }: { to: string; children: ReactNode }) {
	return (
		<Link to={to} className="mb-5 inline-flex items-center gap-1 text-sm text-muted no-underline hover:text-fg">
			<ChevronLeftIcon />
			{children}
		</Link>
	);
}

/* ---------- Alerts ---------- */

export function Alert({ kind = "error", children }: { kind?: "error" | "success"; children: ReactNode }) {
	if (!children) return null;
	const error = kind === "error";
	return (
		<div
			role={error ? "alert" : "status"}
			className={cx(
				"mb-5 flex items-start gap-2.5 rounded-field border px-3.5 py-3 text-sm",
				error
					? "border-red-500/28 bg-red-500/8 text-error-text"
					: "border-green-500/28 bg-green-500/8 text-success-text",
			)}
		>
			{error ? <AlertIcon className="mt-px shrink-0" /> : <CheckIcon className="mt-px shrink-0" />}
			<span>{children}</span>
		</div>
	);
}

/* ---------- Form fields ---------- */

const fieldClass =
	"w-full rounded-field border border-line bg-surface px-3.5 py-3 text-[15px] text-fg placeholder:text-placeholder outline-none transition hover:border-line-hover focus:border-line-hover focus:ring-4 focus:ring-ring";

export function Field({ label, htmlFor, hint, children }: { label: string; htmlFor: string; hint?: string; children: ReactNode }) {
	return (
		<div className="mb-5">
			<label htmlFor={htmlFor} className="mb-2 block text-sm font-medium">
				{label}
			</label>
			{children}
			{hint && <p className="mt-2 text-[0.8rem] text-muted">{hint}</p>}
		</div>
	);
}

export function Input(props: ComponentProps<"input">) {
	return <input {...props} className={cx(fieldClass, props.className)} />;
}

export function Textarea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
	return <textarea {...props} className={cx(fieldClass, "min-h-24 resize-y leading-relaxed", props.className)} />;
}

export function PasswordInput(props: Omit<InputHTMLAttributes<HTMLInputElement>, "type">) {
	const [visible, setVisible] = useState(false);
	return (
		<div className="relative">
			<Input {...props} type={visible ? "text" : "password"} className="pr-11" />
			<button
				type="button"
				onClick={() => setVisible((v) => !v)}
				aria-label={visible ? "Hide password" : "Show password"}
				aria-pressed={visible}
				className="absolute inset-y-0 right-0 flex w-11 cursor-pointer items-center justify-center text-muted hover:text-fg"
			>
				{visible ? <EyeOffIcon /> : <EyeIcon />}
			</button>
		</div>
	);
}

/* ---------- Buttons ---------- */

type Variant = "primary" | "outline" | "danger";

function buttonClass(variant: Variant, small?: boolean, block?: boolean) {
	return cx(
		"inline-flex cursor-pointer items-center justify-center gap-2 rounded-field font-medium no-underline transition disabled:cursor-not-allowed disabled:opacity-60",
		small ? "px-3 py-1.5 text-[13px]" : "px-4 py-3 text-[15px]",
		block && "w-full",
		variant === "primary" && "bg-primary text-primary-fg hover:bg-primary-hover",
		variant === "outline" && "border border-line bg-transparent text-fg hover:border-line-hover hover:bg-surface",
		variant === "danger" && "border border-line bg-transparent text-error-text hover:border-red-500/40 hover:bg-red-500/8",
	);
}

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; small?: boolean; block?: boolean };

export function Button({ variant = "primary", small, block, className, ...props }: ButtonProps) {
	return <button {...props} className={cx(buttonClass(variant, small, block), className)} />;
}

export function ButtonLink({ variant = "outline", small, className, ...props }: LinkProps & { variant?: Variant; small?: boolean }) {
	return <Link {...props} className={cx(buttonClass(variant, small), className)} />;
}
