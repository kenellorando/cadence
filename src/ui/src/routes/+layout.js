// The app is a single page prerendered to static files at build time, which the
// Go server serves out of public/. Prerendering with SSR on means the shipped
// HTML contains the real markup rather than an empty shell; the live values are
// filled in on the client once the API responds.
export const prerender = true;
