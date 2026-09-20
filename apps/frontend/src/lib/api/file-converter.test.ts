import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { FakeEventSource } from '$lib/testing/fake-event-source';
import {
	downloadHref,
	fetchJobStatus,
	startConversion,
	uploadFile,
	watchJob
} from './file-converter';

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

describe('downloadHref', () => {
	it('points at the download route of the job', () => {
		expect(downloadHref('abc')).toBe('/api/file-converter/files/abc/download');
	});

	it('URL-encodes the handle', () => {
		expect(downloadHref('a/b')).toBe('/api/file-converter/files/a%2Fb/download');
	});
});

describe('watchJob', () => {
	beforeEach(() => {
		FakeEventSource.reset();
		vi.stubGlobal('EventSource', FakeEventSource);
	});

	it('opens the event stream of the job, with the handle URL-encoded', () => {
		watchJob('a/b', vi.fn());

		expect(FakeEventSource.instances.map((es) => es.url)).toEqual([
			'/api/file-converter/files/a%2Fb/events'
		]);
	});

	it('passes the update without the server envelope to the callback', () => {
		const onUpdate = vi.fn();
		watchJob('abc', onUpdate);

		FakeEventSource.latest().emit('{"data":{"status":"processing"}}');

		expect(onUpdate).toHaveBeenCalledExactlyOnceWith({ status: 'processing' });
	});

	it('passes the error message of a failed job', () => {
		const onUpdate = vi.fn();
		watchJob('abc', onUpdate);

		FakeEventSource.latest().emit('{"data":{"status":"failed","error":"boom"}}');

		expect(onUpdate).toHaveBeenCalledExactlyOnceWith({ status: 'failed', error: 'boom' });
	});

	it.each(['uploaded', 'pending', 'processing'])(
		'keeps the stream open while the job is %s',
		(status) => {
			watchJob('abc', vi.fn());

			FakeEventSource.latest().emit(`{"data":{"status":"${status}"}}`);

			expect(FakeEventSource.latest().readyState).not.toBe(FakeEventSource.CLOSED);
		}
	);

	it.each(['done', 'failed'])(
		'delivers the %s update, then closes the stream so the browser does not reconnect',
		(status) => {
			const onUpdate = vi.fn();
			watchJob('abc', onUpdate);

			FakeEventSource.latest().emit(`{"data":{"status":"${status}"}}`);

			expect(onUpdate).toHaveBeenCalledExactlyOnceWith({ status });
			expect(FakeEventSource.latest().readyState).toBe(FakeEventSource.CLOSED);
		}
	);

	it('reports a failure when the browser gives up on the connection', () => {
		const onUpdate = vi.fn();
		watchJob('abc', onUpdate);

		FakeEventSource.latest().fail(true);

		expect(onUpdate).toHaveBeenCalledExactlyOnceWith({
			status: 'failed',
			error: expect.any(String)
		});
	});

	it('stays silent on a dropped connection that the browser retries', () => {
		const onUpdate = vi.fn();
		watchJob('abc', onUpdate);

		FakeEventSource.latest().fail(false);

		expect(onUpdate).not.toHaveBeenCalled();
	});

	it('closes the stream when the returned function is called', () => {
		const stop = watchJob('abc', vi.fn());

		stop();

		expect(FakeEventSource.latest().readyState).toBe(FakeEventSource.CLOSED);
	});

	it.each(['not json', '{}', '{"data":null}', 'null'])(
		'reports a failure and closes the stream on the malformed message %s',
		(raw) => {
			const onUpdate = vi.fn();
			watchJob('abc', onUpdate);

			expect(() => FakeEventSource.latest().emit(raw)).not.toThrow();

			expect(onUpdate).toHaveBeenCalledExactlyOnceWith({
				status: 'failed',
				error: expect.any(String)
			});
			expect(FakeEventSource.latest().readyState).toBe(FakeEventSource.CLOSED);
		}
	);
});
