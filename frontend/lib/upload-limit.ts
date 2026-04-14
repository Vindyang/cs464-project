const parsedLimit = Number(process.env.NEXT_PUBLIC_MAX_UPLOAD_MB ?? "30");

export const MAX_UPLOAD_MB =
	Number.isFinite(parsedLimit) && parsedLimit > 0 ? Math.floor(parsedLimit) : 30;
export const MAX_UPLOAD_BYTES = MAX_UPLOAD_MB * 1024 * 1024;
export const UPLOAD_LIMIT_LABEL = `${MAX_UPLOAD_MB} MB`;
