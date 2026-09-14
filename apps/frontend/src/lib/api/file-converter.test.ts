import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fetchJobStatus, startConversion, uploadFile } from './file-converter';

const jsonResponse = (status: number, body: unknown) =>
	new Response(JSON.stringify(body), { status });

class FakeXhr {
	status = 0;
	response = '';
	upload = { addEventListener: vi.fn() };
	open = vi.fn();
	send = vi.fn();
	onload: (() => void) | null = null;
	onerror: (() => void) | null = null;

	respond(status: number, response: string) {
		this.status = status;
		this.response = response;
		this.onload?.();
	}

	error() {
		this.onerror?.();
	}
}

let fakeXhr: FakeXhr;

beforeEach(() => {
	vi.stubGlobal('fetch', vi.fn());
	fakeXhr = new FakeXhr();
	vi.stubGlobal(
		'XMLHttpRequest',
		vi.fn(function () {
			return fakeXhr;
		})
	);
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('uploadFile', () => {
	it('resolves with the uploaded file data on success', async () => {
		const file = new File(['content'], 'report.docx');

		const pending = uploadFile(file);
		fakeXhr.respond(
			200,
			JSON.stringify({
				data: { handle: 'abc', detectedSource: 'docx', targets: ['pdf', 'txt'] }
			})
		);

		const result = await pending;

		expect(result).toEqual({
			success: true,
			data: { handle: 'abc', detectedSource: 'docx', targets: ['pdf', 'txt'] }
		});
		expect(fakeXhr.open).toHaveBeenCalledWith('POST', '/api/file-converter/files', true);
	});

	it('resolves with the problem details on failure', async () => {
		const file = new File(['content'], 'report.exe');

		const pending = uploadFile(file);
		fakeXhr.respond(
			415,
			JSON.stringify({
				type: 'about:blank',
				title: 'Unsupported Media Type',
				status: 415,
				detail: 'File type not supported'
			})
		);

		await expect(pending).rejects.toEqual({
			success: false,
			type: 'about:blank',
			title: 'Unsupported Media Type',
			status: 415,
			detail: 'File type not supported'
		});
	});

	it('resolves with a network error when the request itself fails', async () => {
		const file = new File(['content'], 'report.docx');

		const pending = uploadFile(file);
		fakeXhr.status = 0;
		fakeXhr.error();

		await expect(pending).rejects.toEqual({
			success: false,
			type: 'about:blank',
			title: 'Network Error',
			status: 0,
			detail: 'The request failed or the response could not be read.'
		});
	});

	it('resolves with a network error when the response body is not JSON', async () => {
		const file = new File(['content'], 'report.docx');

		const pending = uploadFile(file);
		fakeXhr.respond(502, '<html>Bad Gateway</html>');

		await expect(pending).rejects.toEqual({
			success: false,
			type: 'about:blank',
			title: 'Network Error',
			status: 502,
			detail: 'The request failed or the response could not be read.'
		});
	});
});

describe('startConversion', () => {
	it('resolves with no data on success', async () => {
		vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 202 }));

		const result = await startConversion('abc', 'pdf');

		expect(result).toEqual({ success: true, data: null });
		expect(fetch).toHaveBeenCalledWith(
			'/api/file-converter/files/abc/convert?target=pdf',
			expect.objectContaining({ method: 'POST' })
		);
	});

	it('resolves with the problem details on failure', async () => {
		vi.mocked(fetch).mockResolvedValue(
			jsonResponse(409, {
				type: 'about:blank',
				title: 'Conflict',
				status: 409,
				detail: 'Conversion already in progress'
			})
		);

		const result = await startConversion('abc', 'pdf');

		expect(result).toEqual({
			success: false,
			type: 'about:blank',
			title: 'Conflict',
			status: 409,
			detail: 'Conversion already in progress'
		});
	});

	it('resolves with a network error when fetch itself fails', async () => {
		vi.mocked(fetch).mockRejectedValue(new TypeError('Failed to fetch'));

		const result = await startConversion('abc', 'pdf');

		expect(result).toEqual({
			success: false,
			type: 'about:blank',
			title: 'Network Error',
			status: 0,
			detail: 'The request failed or the response could not be read.'
		});
	});

	it('resolves with a network error when the failure body is not JSON', async () => {
		vi.mocked(fetch).mockResolvedValue(new Response('<html>Bad Gateway</html>', { status: 502 }));

		const result = await startConversion('abc', 'pdf');

		expect(result).toEqual({
			success: false,
			type: 'about:blank',
			title: 'Network Error',
			status: 502,
			detail: 'The request failed or the response could not be read.'
		});
	});
});

describe('fetchJobStatus', () => {
	it('resolves with the job status on success', async () => {
		vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { data: { status: 'done' } }));

		const result = await fetchJobStatus('abc');

		expect(result).toEqual({ success: true, data: { status: 'done' } });
		expect(fetch).toHaveBeenCalledWith(
			'/api/file-converter/files/abc',
			expect.objectContaining({ method: 'GET' })
		);
	});

	it('resolves with the problem details on failure', async () => {
		vi.mocked(fetch).mockResolvedValue(
			jsonResponse(404, {
				type: 'about:blank',
				title: 'Not Found',
				status: 404,
				detail: 'No such job'
			})
		);

		const result = await fetchJobStatus('abc');

		expect(result).toEqual({
			success: false,
			type: 'about:blank',
			title: 'Not Found',
			status: 404,
			detail: 'No such job'
		});
	});

	it('resolves with a network error when fetch itself fails', async () => {
		vi.mocked(fetch).mockRejectedValue(new TypeError('Failed to fetch'));

		const result = await fetchJobStatus('abc');

		expect(result).toEqual({
			success: false,
			type: 'about:blank',
			title: 'Network Error',
			status: 0,
			detail: 'The request failed or the response could not be read.'
		});
	});
});
