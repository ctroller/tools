import type { ApiResponse, ResponseError } from '$lib/response-types';

export interface UploadResult {
	handle: string;
	detectedSource: string;
	targets: string[];
	name: string;
}

export interface JobStatusResult {
	status: 'uploaded' | 'pending' | 'processing' | 'failed' | 'done';
	error?: string;
}

const BASE_URL = '/api/file-converter';

export const uploadFile = async (file: File): Promise<ApiResponse<UploadResult>> => {
	const formData = new FormData();
	formData.append('file', file);

	const response = await toApiResponse<UploadResult>(
		fetch(`${BASE_URL}/files`, {
			method: 'POST',
			body: formData
		})
	);

	if (!response.success) {
		return response;
	}

	return { ...response, data: { ...response.data, name: file.name } };
};

export const startConversion = async (
	handle: string,
	target: string
): Promise<ApiResponse<null>> => {
	return toEmptyApiResponse(
		fetch(
			`${BASE_URL}/files/${encodeURIComponent(handle)}/convert?target=${encodeURIComponent(target)}`,
			{
				method: 'POST'
			}
		)
	);
};

export const fetchJobStatus = async (handle: string): Promise<ApiResponse<JobStatusResult>> => {
	return toApiResponse(
		fetch(`${BASE_URL}/files/${encodeURIComponent(handle)}`, {
			method: 'GET'
		})
	);
};

async function toApiResponse<T>(res: Promise<Response>): Promise<ApiResponse<T>> {
	const r = await res.catch(() => undefined);
	if (!r) {
		return networkError();
	}

	const body = await r.json().catch(() => undefined);
	if (!body) {
		return networkError(r.status);
	}

	return { ...body, success: r.ok };
}

async function toEmptyApiResponse(res: Promise<Response>): Promise<ApiResponse<null>> {
	const r = await res.catch(() => undefined);
	if (!r) {
		return networkError();
	}
	if (r.ok) {
		return { success: true, data: null };
	}

	const body = await r.json().catch(() => undefined);
	return body ? { ...body, success: false } : networkError(r.status);
}

function networkError(status = 0): ResponseError {
	return {
		success: false,
		type: 'about:blank',
		title: 'Network Error',
		status,
		detail: 'The request failed or the response could not be read.'
	};
}
