import {
	type NetworkCall,
	jsonDeleteWithoutBodyWithoutResponse,
	jsonGetWithoutBodyWithResponse,
	jsonPostWithBodyWithResponse
} from '$lib/utils/networking';
import { categoryStore } from '$lib/stores/category-store.svelte';
import type { CreateNewCategoryRequest } from '$lib/types/requests/create-new-category-request';
import { CategoryResponseSchema, type CategoryResponse } from '$lib/types/responses/category-response';
import {
	CreateCategoryResponseSchema,
	type CreateCategoryResponse
} from '$lib/types/responses/create-category-response';
import { z } from 'zod';

let loading = $state(false);

async function fetchCategories(): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<CategoryResponse[]> = await jsonGetWithoutBodyWithResponse(
			'/api/categories',
			z.array(CategoryResponseSchema)
		);
		if (networkCall.status === 'success') categoryStore.setCategories(networkCall.result);
	} finally {
		loading = false;
	}
}

async function createCategory(request: CreateNewCategoryRequest): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<CreateCategoryResponse> = await jsonPostWithBodyWithResponse(
			'/api/categories',
			request,
			CreateCategoryResponseSchema
		);
		if (networkCall.status === 'success') categoryStore.addCategory({ ...request, id: networkCall.result.id });
	} finally {
		loading = false;
	}
}

async function deleteCategory(id: number): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<void> = await jsonDeleteWithoutBodyWithoutResponse(`/api/categories/${id}`);
		if (networkCall.status === 'success') categoryStore.removeCategory(id);
	} finally {
		loading = false;
	}
}

export const categoryService = {
	get loading(): boolean {
		return loading;
	},
	fetchCategories,
	createCategory,
	deleteCategory
};
