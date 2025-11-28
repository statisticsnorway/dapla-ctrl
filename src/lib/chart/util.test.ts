import { truncateString } from './util';

describe('chart/util', () => {
	describe('truncateString', () => {
		test('returns string unchanged if shorter than limit', () => {
			expect(truncateString('hello', 10)).toBe('hello');
			expect(truncateString('test', 4)).toBe('test');
		});

		test('returns string unchanged if equal to limit', () => {
			expect(truncateString('hello', 5)).toBe('hello');
		});

		test('truncates string and adds ellipsis if longer than limit', () => {
			expect(truncateString('hello world', 8)).toBe('hello...');
			expect(truncateString('this is a long string', 10)).toBe('this is...');
		});

		test('handles exact boundary case', () => {
			// When num=6, we can fit 3 chars + "..." (total 6)
			expect(truncateString('hello world', 6)).toBe('hel...');
		});

		test('handles minimum truncation length', () => {
			// num=4 means 1 char + "..." (4 total)
			expect(truncateString('hello', 4)).toBe('h...');
		});

		test('handles very short limit', () => {
			expect(truncateString('hello', 3)).toBe('...');
		});

		test('handles empty string', () => {
			expect(truncateString('', 5)).toBe('');
		});

		test('handles single character', () => {
			expect(truncateString('a', 5)).toBe('a');
			expect(truncateString('a', 1)).toBe('a');
		});

		test('preserves special characters', () => {
			expect(truncateString('hello@world.com', 10)).toBe('hello@w...');
			expect(truncateString('test-app-name', 8)).toBe('test-...');
		});
	});
});
