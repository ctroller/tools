/**
 * API Response Types - RFC 9457 Compliant
 *
 * This module provides TypeScript interfaces for standardizing REST API responses
 * following RFC 9457 (Problem Details for HTTP APIs).
 *
 * @see https://www.rfc-editor.org/rfc/rfc9457.html
 * @see https://www.rfc-editor.org/rfc/rfc8288.html
 */

// ------------------------------------------------------------------------------
// BASE
// ------------------------------------------------------------------------------

/**
 * Base metadata record for API responses.
 *
 * A generic key-value store for additional metadata fields.
 */
export type BaseMeta = Record<string, unknown>;

/**
 * Web Link representation following RFC 8288.
 *
 * Used for HATEOAS-style navigation and resource relationships.
 *
 * @see https://www.rfc-editor.org/rfc/rfc8288.html
 */
export interface Link {
	/**
	 * The link's target URI.
	 *
	 * @example "https://api.example.com/users?page=2"
	 */
	href: string;

	/**
	 * The relation type of the link.
	 *
	 * @example "next", "prev", "self"
	 */
	rel?: string;

	/**
	 * Human-readable description of the link.
	 */
	title?: string;

	/**
	 * Media type hint for the target resource.
	 *
	 * @example "application/json"
	 */
	type?: string;

	/**
	 * Additional link attributes.
	 */
	[key: string]: unknown;
}

// ------------------------------------------------------------------------------
// PAGINATION
// ------------------------------------------------------------------------------

/**
 * Union type of supported pagination strategies.
 */
export type PaginationType = 'offset' | 'cursor' | 'links';

/**
 * Offset-based pagination (traditional page numbers).
 *
 * Best for small to medium datasets where total count is needed.
 *
 * @example
 * ```json
 * {
 *   "type": "offset",
 *   "total": 156,
 *   "page": 2,
 *   "pageSize": 20,
 *   "pageCount": 8
 * }
 * ```
 */
export interface PaginationOffset {
	/**
	 * Discriminator for pagination type.
	 */
	type: 'offset';

	/**
	 * Total number of items across all pages.
	 */
	total: number;

	/**
	 * Current page number (1-indexed).
	 */
	page: number;

	/**
	 * Number of items per page.
	 */
	pageSize: number;

	/**
	 * Total number of pages available.
	 */
	pageCount: number;
}

/**
 * Cursor-based pagination (for large datasets).
 *
 * Provides stable pagination without counting total items.
 * Best for infinite scroll and real-time data.
 *
 * @example
 * ```json
 * {
 *   "type": "cursor",
 *   "pageSize": 20,
 *   "endCursor": "eyJpZCI6MTAwfQ==",
 *   "hasNextPage": true,
 *   "hasPreviousPage": false
 * }
 * ```
 */
export interface PaginationCursor {
	/**
	 * Discriminator for pagination type.
	 */
	type: 'cursor';

	/**
	 * Number of items per page.
	 */
	pageSize: number;

	/**
	 * Cursor pointing to the start of the current page.
	 * Base64-encoded or opaque string.
	 */
	startCursor?: string;

	/**
	 * Cursor pointing to the end of the current page.
	 * Used to fetch the next page.
	 */
	endCursor?: string;

	/**
	 * Whether there are more items after the current page.
	 */
	hasNextPage: boolean;

	/**
	 * Whether there are items before the current page.
	 */
	hasPreviousPage: boolean;
}

/**
 * Link-based pagination (HATEOAS/Hypermedia approach).
 *
 * Provides navigation through hypermedia links without exposing
 * pagination implementation details.
 *
 * @example
 * ```json
 * {
 *   "type": "links",
 *   "self": { "href": "/users?page=2" },
 *   "first": { "href": "/users?page=1" },
 *   "prev": { "href": "/users?page=1" },
 *   "next": { "href": "/users?page=3" },
 *   "last": { "href": "/users?page=10" }
 * }
 * ```
 */
export interface PaginationLinks {
	/**
	 * Discriminator for pagination type.
	 */
	type: 'links';

	/**
	 * Link to the current page.
	 */
	self?: Link;

	/**
	 * Link to the first page.
	 */
	first?: Link;

	/**
	 * Link to the last page.
	 */
	last?: Link;

	/**
	 * Link to the previous page.
	 */
	prev?: Link;

	/**
	 * Link to the next page.
	 */
	next?: Link;

	/**
	 * Link to related resources.
	 */
	related?: Link;
}

/**
 * Discriminated union of all pagination strategies.
 *
 * Use the `type` field to determine which pagination strategy is in use.
 */
export type Pagination = PaginationOffset | PaginationCursor | PaginationLinks;

// ------------------------------------------------------------------------------
// API RESPONSE
// ------------------------------------------------------------------------------

/**
 * RFC 9457 Problem Details for HTTP APIs.
 *
 * Standard format for machine-readable error responses in HTTP APIs.
 * Must be served with `Content-Type: application/problem+json`.
 *
 * @see https://www.rfc-editor.org/rfc/rfc9457.html
 *
 * @example
 * ```json
 * {
 *   "type": "https://api.example.com/errors/validation-error",
 *   "title": "Validation Error",
 *   "status": 400,
 *   "detail": "The request body contains invalid data",
 *   "instance": "/users/create"
 * }
 * ```
 */
export interface ProblemDetails {
	/**
	 * A URI reference that identifies the problem type.
	 *
	 * When dereferenced, it should provide human-readable documentation.
	 *
	 * @default "about:blank"
	 * @example "https://api.example.com/errors/not-found"
	 */
	type: string;

	/**
	 * A short, human-readable summary of the problem type.
	 *
	 * Should be the same for all occurrences of the same problem type.
	 *
	 * @example "Resource Not Found", "Validation Error"
	 */
	title: string;

	/**
	 * The HTTP status code generated by the origin server for this occurrence of the problem.
	 *
	 * @example 404, 400, 500
	 */
	status: number;

	/**
	 * A human-readable explanation specific to this occurrence of the problem.
	 *
	 * Should focus on helping the client correct the problem.
	 *
	 * @example "The user with ID 12345 does not exist"
	 */
	detail?: string;

	/**
	 * A URI reference that identifies the specific occurrence of the problem.
	 *
	 * It may or may not yield further information if dereferenced.
	 *
	 * @example "/users/12345/update"
	 */
	instance?: string;

	/**
	 * Additional members to provide more context about the problem.
	 *
	 * Can include application-specific fields like trace IDs, support info, etc.
	 */
	[k: string]: unknown;
}

/**
 * Individual error detail for validation or business logic errors.
 *
 * Used when a request contains multiple errors that should be reported together.
 *
 * @example
 * ```json
 * {
 *   "message": "Email address is not valid",
 *   "code": "INVALID_EMAIL_FORMAT",
 *   "parameter": "email",
 *   "rejectedValue": "not-an-email"
 * }
 * ```
 */
export interface ErrorDetail {
	/**
	 * Human-readable error message.
	 */
	message: string;

	/**
	 * Application-specific error code for programmatic handling.
	 *
	 * @example "VALIDATION_ERROR", "DUPLICATE_EMAIL"
	 */
	code?: string;

	/**
	 * Name of the query or path parameter that caused the error.
	 *
	 * @example "userId", "startDate"
	 */
	parameter?: string;

	/**
	 * Name of the HTTP header that caused the error.
	 *
	 * @example "Authorization", "Content-Type"
	 */
	header?: string;

	/**
	 * The value that was rejected (useful for debugging).
	 *
	 * @example "invalid-uuid-format"
	 */
	rejectedValue?: unknown;

	/**
	 * Additional error-specific fields.
	 */
	[k: string]: unknown;
}

/**
 * Base interface for all API responses.
 *
 * Contains common fields shared across success and error responses.
 */
export interface ResponseBase {
	/**
	 * Indicates whether the request was successful.
	 */
	success: boolean;

	/**
	 * Unique identifier for tracing the request across systems.
	 *
	 * @example "req_1a2b3c4d5e6f"
	 */
	requestId?: string;

	/**
	 * ISO 8601 timestamp of when the response was generated.
	 *
	 * @example "2025-09-30T14:30:00.000Z"
	 */
	timestamp?: string;

	/**
	 * Additional base response fields.
	 */
	[k: string]: unknown;
}

/**
 * Standard successful API response.
 *
 * @template T - The type of data returned in the response
 *
 * @example
 * ```json
 * {
 *   "success": true,
 *   "data": { "id": 1, "name": "John Doe" },
 *   "requestId": "req_abc123",
 *   "timestamp": "2025-09-30T14:30:00.000Z"
 * }
 * ```
 */
export interface ResponseSuccess<T = unknown> extends ResponseBase {
	/**
	 * Indicates the request was successful.
	 */
	success: true;

	/**
	 * The response payload.
	 */
	data: T;

	/**
	 * Optional metadata about the response.
	 */
	meta?: BaseMeta;

	/**
	 * Optional pagination information for list responses.
	 */
	pagination?: Pagination;
}

/**
 * Standard error API response following RFC 9457.
 *
 * Should be served with `Content-Type: application/problem+json`.
 *
 * @example
 * ```json
 * {
 *   "success": false,
 *   "type": "https://api.example.com/errors/validation",
 *   "title": "Validation Error",
 *   "status": 400,
 *   "detail": "Request validation failed",
 *   "errors": [
 *     {
 *       "message": "Email is required",
 *       "code": "REQUIRED_FIELD",
 *       "parameter": "email"
 *     }
 *   ]
 * }
 * ```
 */
export interface ResponseError extends ResponseBase, ProblemDetails {
	/**
	 * Indicates the request failed.
	 */
	success: false;

	/**
	 * Array of individual error details.
	 *
	 * Useful for reporting multiple validation errors or business rule violations.
	 */
	errors?: ErrorDetail[];
}

/**
 * Discriminated union of all possible API responses.
 *
 * Use the `success` field to determine response type at runtime.
 *
 * @template T - The type of data in successful responses
 *
 * @example
 * ```typescript
 * const response: ApiResponse<User> = await fetchUser(id)
 *
 * if (response.success) {
 *   console.log(response.data.name) // TypeScript knows data exists
 * } else {
 *   console.error(response.title, response.errors)
 * }
 * ```
 */
export type ApiResponse<T = unknown> = ResponseSuccess<T> | ResponseError;
