import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { PAGE_SIZE } from "./paginationUtils";

export default function Pagination({ currentPage, totalItems, onPageChange, pageSize = PAGE_SIZE }) {
  const totalPages = Math.ceil(totalItems / pageSize);
  if (totalPages <= 1) return null;

  const safePage = Math.min(Math.max(currentPage, 1), totalPages);
  const pages = Array.from({ length: totalPages }, (_, index) => index + 1);

  return (
    <nav className="system-pagination" aria-label="เปลี่ยนหน้ารายการ">
      <button type="button" onClick={() => onPageChange(1)} disabled={safePage === 1} aria-label="หน้าแรก">
        <ChevronsLeft size={18} aria-hidden="true" /> <span className="pagination-label">หน้าแรก</span>
      </button>
      <button type="button" onClick={() => onPageChange(safePage - 1)} disabled={safePage === 1} aria-label="หน้าก่อนหน้า">
        <ChevronLeft size={18} aria-hidden="true" /> <span className="pagination-label">ก่อนหน้า</span>
      </button>
      <div className="system-pagination-pages">
        {pages.map((page) => (
          <button
            type="button"
            key={page}
            className={page === safePage ? "is-active" : ""}
            aria-current={page === safePage ? "page" : undefined}
            onClick={() => onPageChange(page)}
          >
            {page}
          </button>
        ))}
      </div>
      <button type="button" onClick={() => onPageChange(safePage + 1)} disabled={safePage === totalPages} aria-label="หน้าถัดไป">
        <span className="pagination-label">ถัดไป</span> <ChevronRight size={18} aria-hidden="true" />
      </button>
      <button type="button" onClick={() => onPageChange(totalPages)} disabled={safePage === totalPages} aria-label="หน้าสุดท้าย">
        <span className="pagination-label">หน้าสุดท้าย</span> <ChevronsRight size={18} aria-hidden="true" />
      </button>
    </nav>
  );
}
