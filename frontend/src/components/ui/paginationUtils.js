export const PAGE_SIZE = 10;
export const REPAIR_PAGE_SIZE = 4;

export function getPageItems(items, currentPage, pageSize = PAGE_SIZE) {
  const safePage = Math.min(Math.max(currentPage, 1), Math.max(Math.ceil(items.length / pageSize), 1));
  const start = (safePage - 1) * pageSize;
  return items.slice(start, start + pageSize);
}
