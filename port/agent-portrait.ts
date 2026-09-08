/** Stable, static pen-and-ink portraits. Only seeded numeric geometry enters the SVG. */
export function agentPortrait(seed: string): string {
	let hash = 2166136261;
	for (const char of seed) hash = Math.imul(hash ^ char.codePointAt(0)!, 16777619) >>> 0;
	const initial = hash;
	const next = (max: number) => {
		hash ^= hash << 13; hash ^= hash >>> 17; hash ^= hash << 5;
		return (hash >>> 0) % max;
	};
	const ink = '#292820';
	const paper = ['#d6d0bb', '#ded7c5', '#cfcbb8', '#d8d0be'][initial % 4];
	// Wardrobe and backdrop use separate bits, preserving every existing face and hair stroke.
	const wardrobe = Math.imul(initial ^ 0x6d2b79f5, 0x45d9f3b) >>> 0;
	const outfit = wardrobe % 6;
	const backdrop = ['#d6d0bb', '#c9b68f', '#c5a296', '#b9a69b', '#b8b69b', '#c3abab'][(wardrobe >>> 8) % 6];
	const cloth = ['#3c3a30', '#484234', '#454638', '#51423a'][(wardrobe >>> 16) % 4];
	const faceWidth = 11 + next(4);
	const chin = 44 + next(6);
	const eyeY = 27 + next(3);
	const nose = 34 + next(4);
	const hair = next(5);
	const accessory = next(6);
	const coat = next(3);
	const left = 32 - faceWidth, right = 32 + faceWidth;
	const strokes: string[] = [];
	const path = (d: string, fill = 'none', width = .8, color = ink) => `<path d="${d}" fill="${fill}" stroke="${color}" stroke-width="${width}" stroke-linecap="round" stroke-linejoin="round"/>`;

	// Uneven parallel strokes give the backdrop and cloth a dry engraved finish.
	for (let i = 0; i < 18; i++) {
		const x = 4 + next(57), y = 4 + next(62);
		strokes.push(path(`M${x} ${y}l${1 + next(3)} ${next(3) - 1}`, 'none', .25, '#9b9584'));
	}
	// Consume the original hatch samples in the original order so head generation never changes.
	const hatchLengths = Array.from({ length: 16 }, () => 7 + next(6));
	const shirtFill = outfit === 3 ? paper : outfit === 0 ? ink : cloth;
	strokes.push(path('M5 72L9 58Q16 51 26 49L38 49Q49 53 55 59L61 72Z', shirtFill));
	for (let i = 0; i < 16; i++) {
		const x = 9 + i * 3;
		strokes.push(path(`M${x} 71l${coat === 1 ? 4 : -2} -${hatchLengths[i]}`, 'none', .45, outfit === 3 ? '#8b8471' : paper));
	}
	strokes.push(path('M27 43L26 51L32 58L39 50L37 43', paper));
	if (outfit === 0) {
		// The original evening shirt and bow tie remain in the wardrobe.
		strokes.push(path('M25 49L18 54L25 65L31 56M39 49L46 54L38 64L33 56', paper, .9));
		strokes.push(path('M29 56l-5 2 5 4 3-3 3 3 5-4-6-2Z', ink));
		strokes.push(path('M32 61v11', 'none', .6, paper));
	} else if (outfit === 1) {
		// A narrow tie and buttoned waistcoat.
		strokes.push(path('M26 49L19 53L28 65L32 59L36 65L45 53L38 49L32 54Z', paper));
		strokes.push(path('M30 54h4l-1 5 2 11-3 2-3-2 2-11Z', ink, .55));
		strokes.push(path('M18 54l7 9-4 2 9 7M46 54l-7 9 4 2-9 7', 'none', .6, paper));
	} else if (outfit === 2) {
		// A high ribbed pullover, without a collar or tie.
		strokes.push(path('M25 49Q32 52 39 49L40 57Q32 61 24 57Z', cloth));
		for (let x = 26; x <= 38; x += 2) strokes.push(path(`M${x} 51v6`, 'none', .4, paper));
		strokes.push(path('M16 57q-2 6-1 14M48 57q2 6 1 14', 'none', .7, paper));
	} else if (outfit === 3) {
		// An open ivory shirt with dark braces.
		strokes.push(path('M26 49L22 52L27 60L32 54L37 60L42 52L38 49L32 54Z', paper));
		strokes.push(path('M17 54L14 72h4l3-20M43 52l3 20h4l-3-18', ink, .6));
		strokes.push(path('M32 56v16', 'none', .55));
	} else if (outfit === 4) {
		// A round-neck knit with a lightly crosshatched chest.
		strokes.push(path('M23 51Q32 64 41 51L44 53Q32 69 20 53Z', cloth));
		strokes.push(path('M23 54q9 12 18 0', 'none', .65, paper));
		for (let y = 63; y <= 71; y += 3) strokes.push(path(`M13 ${y}h38`, 'none', .35, paper));
	} else {
		// A soft shawl collar and double-breasted cardigan.
		strokes.push(path('M26 49Q18 50 19 57L31 69L30 58Z', paper));
		strokes.push(path('M38 49Q46 50 45 57L33 69L34 58Z', paper));
		strokes.push(path('M31 64v8M33 64v8', 'none', .6, paper));
	}
	if ([0, 3, 5].includes(outfit)) {
		const buttonInk = outfit === 3 ? ink : paper;
		strokes.push(`<circle cx="35" cy="66" r=".8" fill="${buttonInk}"/><circle cx="35" cy="71" r=".8" fill="${buttonInk}"/>`);
	}
	strokes.push(path(`M${left} 28Q${left - 5} 23 ${left - 3} 34L${left + 1} 38M${right} 28Q${right + 5} 24 ${right + 3} 34L${right - 1} 38`, paper));
	strokes.push(path(`M${left} 24Q${left - 1} 10 32 10Q${right + 2} 11 ${right} 25L${right - 1} 37Q${right - 3} ${chin} 32 ${chin + 1}Q${left + 3} ${chin - 1} ${left + 1} 36Z`, paper, 1));
	for (let i = 0; i < 5; i++) strokes.push(path(`M${left + 1 + i * .6} ${32 + i}l2 4`, 'none', .35));
	for (let i = 0; i < 4; i++) strokes.push(path(`M${right - 3 + i * .6} ${30 + i}l-1 5`, 'none', .3));
	// Small hooded eyes, angular noses and restrained expressions, rather than emoji faces.
	strokes.push(path(`M${left + 3} ${eyeY - 3}l6 -1M${right - 9} ${eyeY - 4}l6 2`, 'none', 1));
	strokes.push(path(`M${left + 3} ${eyeY}q3 -2 6 0M${right - 9} ${eyeY}q3 -2 6 0`));
	strokes.push(`<ellipse cx="${left + 6}" cy="${eyeY}" rx=".8" ry="1.1" fill="${ink}"/><ellipse cx="${right - 6}" cy="${eyeY}" rx=".8" ry="1.1" fill="${ink}"/>`);
	strokes.push(path(`M32 ${eyeY - 2}L29 ${nose}Q31 ${nose + 2} 35 ${nose}M29 ${nose + 5}q4 -1 ${5 + next(3)} 0M31 ${nose + 7}h3`, 'none', .7));
	if (accessory === 0) {
		strokes.push(`<g fill="none" stroke="${ink}" stroke-width=".7"><circle cx="${left + 6}" cy="${eyeY}" r="4.2"/><circle cx="${right - 6}" cy="${eyeY}" r="4.2"/></g>`);
		strokes.push(path(`M${left + 10} ${eyeY}h${2 * faceWidth - 20}`));
	} else if (accessory === 1) {
		strokes.push(path(`M32 ${nose + 2}q-4 -3 -9 3 6 1 9-1 3 3 9 0-5-5-9-2Z`, ink, .5));
	} else if (accessory === 2) {
		strokes.push(path(`M27 ${chin - 2}l5 7 5-7-5 2Z`, ink, .5));
	} else if (accessory === 3) {
		strokes.push(`<circle cx="${right - 6}" cy="${eyeY}" r="4.5" fill="none" stroke="${ink}" stroke-width=".8"/>`);
		strokes.push(path(`M${right - 2} ${eyeY + 2}q9 16 4 25`, 'none', .5));
	}
	if (hair === 0) {
		strokes.push(path(`M${left - 1} 20L${left + 1} 4L${right - 1} 5L${right + 1} 20Z`, ink));
		strokes.push(path(`M${left - 6} 21Q32 17 ${right + 6} 22Q31 25 ${left - 6} 21Z`, ink));
		for (let i = 0; i < 5; i++) strokes.push(path(`M${left + 3 + i * 2} 7l-1 10`, 'none', .35, paper));
		strokes.push(path(`M${left + 1} 17L${right} 18`, 'none', 1.3, paper));
	} else if (hair === 1) {
		strokes.push(path(`M${left - 1} 26Q${left - 8} 8 26 9Q34 3 ${right + 2} 15L${right + 1} 26L${right - 3} 18Q31 16 27 13L${left + 4} 24Z`, ink));
		for (let i = 0; i < 7; i++) strokes.push(path(`M${left + 3 + i * 2} 20q-3 -6 2 -9`, 'none', .4, paper));
	} else if (hair === 2) {
		strokes.push(path(`M${left} 30L${left - 3} 25L${left - 1} 21L${left - 5} 17L${left + 1} 16L${left} 9L${left + 5} 12L${left + 9} 6L32 11L38 7L41 12L${right + 4} 11L${right + 1} 18L${right + 5} 23L${right} 30L${right - 2} 19Q31 21 25 15L${left + 2} 25Z`, ink, .6));
	} else if (hair === 3) {
		strokes.push(path(`M${left - 1} 29Q${left - 7} 10 28 9Q42 5 ${right + 3} 22L${right} 34L${right - 2} 24Q37 16 29 13Q23 20 ${left + 1} 25Z`, ink));
		for (let i = 0; i < 6; i++) strokes.push(path(`M${left + 1} ${13 + i}Q30 ${6 + i} ${right} ${17 + i}`, 'none', .35, paper));
	} else {
		strokes.push(path(`M${left - 1} 33Q${left - 6} 18 ${left + 4} 14L${left + 2} 28L${left + 1} 35M${right} 33Q${right + 6} 19 ${right - 3} 14L${right - 2} 28L${right - 1} 35`, ink));
		strokes.push(path('M29 13q3 -4 5 0M27 15q3 -3 8 0', 'none', .5));
	}
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 72" fill="none" aria-hidden="true"><path fill="${backdrop}" d="M0 0h64v72H0z"/>${strokes.join('')}</svg>`;
}
