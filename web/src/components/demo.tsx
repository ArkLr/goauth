"use client";

import { GlyphMatrix } from "@/components/ui/glyph-matrix";
import * as React from "react";

export default function GlyphMatrixDemo() {
	const [color, setColor] = React.useState("#6B7280");

	return (
		<div className="border-border bg-background relative h-[400px] w-full overflow-hidden rounded-lg border">
			<GlyphMatrix
				glyphs="01.+*/<>="
				cellSize={14}
				mutationRate={0.04}
				interval={90}
				fadeBottom={0.6}
				color={color}
			/>
		</div>
	);
}
