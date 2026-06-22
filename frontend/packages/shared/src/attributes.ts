export const ATTRIBUTE_MIN = 1;

export function clampAttribute(value: number): number {
	return Math.max(ATTRIBUTE_MIN, value);
}

export const ATTRIBUTE_BAR_MAX = 12;
export const ATTRIBUTE_BAR_MIN_PERCENT = 8;

export function attributeBarWidth(value: number): number {
	return Math.min(
		100,
		Math.max(ATTRIBUTE_BAR_MIN_PERCENT, value * (100 / ATTRIBUTE_BAR_MAX)),
	);
}
