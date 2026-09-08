import { describe, it } from 'vitest';

describe('backlog page tests', () => {
	// TODO: Cover navbar actions.

	describe('category operations', () => {
		it.todo('successfully creates and deletes a category');
		it.todo('shows an error toast when deleting a referenced category');
	});

	describe('task creation', () => {
		it.todo('reacts correctly to missing task title');
		it.todo('reacts correctly to too long task title');
		it.todo('reacts correctly to too long task description');
		it.todo('reacts correctly to missing task category');
		it.todo('cancels and closes the create-task dialog');
		it.todo('successfully adds a task');
	});

	describe('task completion', () => {
		it.todo('successfully completes a task without a date');
		it.todo('successfully completes a task with a non-today date');
		it.todo('successfully completes a task dated today');
	});

	describe('task trashing', () => {
		it.todo('successfully trashes a task without a date');
		it.todo('successfully trashes a task with a non-today date');
		it.todo('successfully trashes a task dated today');
	});

	describe('task editing', () => {
		it.todo('reacts correctly to missing task title');
		it.todo('reacts correctly to too long task title');
		it.todo('reacts correctly to too long task description');
		it.todo('reacts correctly to missing task category');
		it.todo('cancels and closes the edit-task dialog');
		it.todo('successfully edits a task without a date');
		it.todo('successfully edits a task with a non-today date');
		it.todo('successfully edits a task dated today');
	});

	describe('drag and drop', () => {
		describe('within the no-date partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('within the non-today-date partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('between different non-today date groups', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('within the today partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('from the no-date partition', () => {
			it.todo('moves a task to the first index of the non-today-date partition');
			it.todo('moves a task to the middle index of the non-today-date partition');
			it.todo('moves a task to the last index of the non-today-date partition');
			it.todo('moves a task to the first index of the today partition');
			it.todo('moves a task to the middle index of the today partition');
			it.todo('moves a task to the last index of the today partition');
		});

		describe('from the non-today-date partition', () => {
			it.todo('moves a task to the first index of the no-date partition');
			it.todo('moves a task to the middle index of the no-date partition');
			it.todo('moves a task to the last index of the no-date partition');
			it.todo('moves a task to the first index of the today partition');
			it.todo('moves a task to the middle index of the today partition');
			it.todo('moves a task to the last index of the today partition');
		});

		describe('from the today partition', () => {
			it.todo('moves a task to the first index of the no-date partition');
			it.todo('moves a task to the middle index of the no-date partition');
			it.todo('moves a task to the last index of the no-date partition');
			it.todo('moves a task to the first index of the non-today-date partition');
			it.todo('moves a task to the middle index of the non-today-date partition');
			it.todo('moves a task to the last index of the non-today-date partition');
		});
	});

	describe('task batch operations', () => {
		it.todo('successfully completes all tasks');
		it.todo('successfully trashes all tasks');
		it.todo('successfully cancels a batch operation');
		it.todo('disables batch actions when no tasks are selected');
		it.todo('disables batch trash when more than 50 tasks are selected');
		it.todo('disables batch completion when more than 50 tasks are selected');
	});
});
