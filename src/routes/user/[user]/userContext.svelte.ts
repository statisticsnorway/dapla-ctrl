import { getContext, setContext } from 'svelte';

const ctxKey = Symbol('user');

class UserContext {
	refetchCounts() {
		this.inventoryFetcher?.();
	}

	inventoryFetcher?: () => void;
}

export function createUserContext() {
	const context = new UserContext();
	setContext(ctxKey, context);
}

export function setInventoryRefetcher(refetcher: () => void) {
	const ctx = getContext<UserContext>(ctxKey);
	ctx.inventoryFetcher = refetcher;
}

export function getUserContext() {
	return getContext<UserContext>(ctxKey);
}
