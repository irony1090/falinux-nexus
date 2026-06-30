
export const findParent = <T extends HTMLElement>(target: T, from: HTMLElement, depth: number = 1): T|null => {
    if (!from?.parentElement) return null;
    if (from.parentElement === target) return from.parentElement as T;

    if (depth <= 0) return null;
    return findParent(target, from.parentElement, depth - 1);
}

// export const findParent