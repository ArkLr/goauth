import type { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement> & { size?: number };

function Icon({ size = 18, children, ...props }: IconProps) {
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth={2}
			strokeLinecap="round"
			strokeLinejoin="round"
			aria-hidden="true"
			{...props}
		>
			{children}
		</svg>
	);
}

export const AlertIcon = (p: IconProps) => (
	<Icon {...p}><circle cx="12" cy="12" r="10" /><path d="M12 8v4M12 16h.01" /></Icon>
);
export const CheckIcon = (p: IconProps) => (
	<Icon {...p}><circle cx="12" cy="12" r="10" /><path d="m8 12 3 3 5-6" /></Icon>
);
export const EyeIcon = (p: IconProps) => (
	<Icon {...p}><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" /><circle cx="12" cy="12" r="3" /></Icon>
);
export const EyeOffIcon = (p: IconProps) => (
	<Icon {...p}>
		<path d="M10.7 5.1A10.4 10.4 0 0 1 12 5c6.5 0 10 7 10 7a17.6 17.6 0 0 1-2.2 3.2" />
		<path d="M6.6 6.6A17.4 17.4 0 0 0 2 12s3.5 7 10 7a9.7 9.7 0 0 0 5.4-1.6" />
		<path d="M9.9 9.9a3 3 0 0 0 4.2 4.2" />
		<path d="m2 2 20 20" />
	</Icon>
);
export const ChevronLeftIcon = (p: IconProps) => <Icon size={16} {...p}><path d="m15 18-6-6 6-6" /></Icon>;
export const UserIcon = (p: IconProps) => (
	<Icon {...p}><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></Icon>
);
export const PlusIcon = (p: IconProps) => <Icon {...p}><path d="M12 5v14M5 12h14" /></Icon>;
export const ListIcon = (p: IconProps) => <Icon {...p}><path d="M4 6h16M4 12h16M4 18h10" /></Icon>;
export const PencilIcon = (p: IconProps) => (
	<Icon size={14} {...p}><path d="M12 20h9" /><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" /></Icon>
);
export const TrashIcon = (p: IconProps) => (
	<Icon size={14} {...p}><path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="m19 6-1 14H6L5 6" /></Icon>
);
export const LogOutIcon = (p: IconProps) => (
	<Icon {...p}><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="m16 17 5-5-5-5" /><path d="M21 12H9" /></Icon>
);
