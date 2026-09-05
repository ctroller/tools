import type {RouteId} from '$app/types';

export type Tool = {
	partNumber: string;
	name: string;
	description: string;
	href: RouteId;
	tags: string[];
};

export const Tools: Tool[] = [
	{
		partNumber: '001',
		name: 'File Converter',
		description: 'Convert files between various formats.',
		href: '/tools/file-converter',
		tags: ['file', 'converter']
	}
];
