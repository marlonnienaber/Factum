import type { CategoryResponse } from '$lib/types/responses/category-response';

let categories: CategoryResponse[] = $state<CategoryResponse[]>([]);

function setCategories(newCategories: CategoryResponse[]): void {
	categories = newCategories;
}

function addCategory(category: CategoryResponse): void {
	categories = [...categories, category];
}

function removeCategory(id: number): void {
	categories = categories.filter((category) => category.id !== id);
}

function reset(): void {
	categories = [];
}

export const categoryStore = {
	get categories() {
		return categories;
	},
	setCategories,
	addCategory,
	removeCategory,
	reset
};
