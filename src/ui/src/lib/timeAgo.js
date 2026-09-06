// Renders a play-history timestamp as a coarse relative time, matching the
// wording the previous UI used.
export function timeAgo(when) {
	const seconds = Math.round((Date.now() - new Date(when).getTime()) / 1000);
	const minute = 60;
	const hour = minute * 60;
	const day = hour * 24;

	if (seconds < 30) return 'just now';
	if (seconds < minute) return `${seconds} seconds ago`;
	if (seconds < 2 * minute) return 'a minute ago';
	if (seconds < hour) return `${Math.floor(seconds / minute)} minutes ago`;
	if (Math.floor(seconds / hour) === 1) return '1 hour ago';
	if (seconds < day) return `${Math.floor(seconds / hour)} hours ago`;
	return 'a while ago';
}
