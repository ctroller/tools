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

export const uploadFile = async (
	file: File,
	progressHandler?: (sent: number, total: number) => void
): Promise<ApiResponse<UploadResult>> => {
	return new Promise((resolve, reject) => {
		const formData = new FormData();
		formData.append('file', file);

		const xhr = new XMLHttpRequest();
		xhr.upload.addEventListener('progress', (e) => {
			if (e.lengthComputable) {
				progressHandler?.(e.loaded, e.total);
			}
		});

		xhr.open('POST', `${BASE_URL}/files`, true);
		xhr.send(formData);

		xhr.onload = () => {
			let r;
			try {
				r = JSON.parse(xhr.response);
			} catch {
				reject(networkError(xhr.status));
				return;
			}
			if (xhr.status >= 200 && xhr.status < 400) {
				resolve({ ...r, success: true } as ApiResponse<UploadResult>);
			} else {
				reject({ ...r, success: false } as ResponseError);
			}
		};

		xhr.onerror = () => {
			reject(networkError(xhr.status));
		};
	});
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

export const deleteJob = async (handle: string): Promise<ApiResponse<null>> => {
	return toEmptyApiResponse(
		fetch(`${BASE_URL}/files/${encodeURIComponent(handle)}`, {
			method: 'DELETE'
		})
	);
};

export const downloadUrl = async (url: string): Promise<ApiResponse<UploadResult>> => {
	return toApiResponse(
		fetch(`${BASE_URL}/download-url?url=${encodeURIComponent(url)}`, {
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
