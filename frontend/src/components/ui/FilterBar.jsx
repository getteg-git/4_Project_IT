import { useEffect, useRef, useState } from "react";
import { Check, SlidersHorizontal, X } from "lucide-react";

export default function FilterBar({ filters }) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);
  const activeCount = filters.filter((filter) => filter.value !== "").length;

  useEffect(() => {
    if (!isOpen) return undefined;

    const handlePointerDown = (event) => {
      if (!containerRef.current?.contains(event.target)) setIsOpen(false);
    };
    const handleKeyDown = (event) => {
      if (event.key === "Escape") setIsOpen(false);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  return (
    <div className={`system-filter ${filters.length === 1 ? "is-compact" : ""}`} ref={containerRef}>
      <button type="button" className={activeCount > 0 ? "system-filter-trigger is-active" : "system-filter-trigger"} onClick={() => setIsOpen((open) => !open)} aria-expanded={isOpen}>
        <SlidersHorizontal size={18} aria-hidden="true" />
        <span className="filter-trigger-label">ฟิลเตอร์</span>
        {activeCount > 0 && <span className="filter-active-count">{activeCount}</span>}
      </button>

      {isOpen && (
        <div className="system-filter-menu" role="dialog" aria-label="เลือกตัวกรอง">
          <div className="system-filter-menu-header">
            <div><strong>ฟิลเตอร์</strong>{filters.length > 1 && <small>เลือกได้มากกว่าหนึ่งเงื่อนไข</small>}</div>
            <button type="button" onClick={() => setIsOpen(false)} aria-label="ปิดเมนูตัวกรอง"><X size={18} aria-hidden="true" /></button>
          </div>
          <div className="system-filter-fields">
            {filters.map((filter) => (
              <label key={filter.id}>
                <span>{filter.label}</span>
                <select value={filter.value} onChange={(event) => filter.onChange(event.target.value)}>
                  <option value="">ทั้งหมด</option>
                  {filter.options.map((option) => {
                    const isOptionObject = typeof option === "object" && option !== null;
                    const value = isOptionObject ? option.value : String(option);
                    const label = isOptionObject ? option.label : String(option);
                    return <option key={value} value={value}>{label}</option>;
                  })}
                </select>
              </label>
            ))}
          </div>
          <div className="system-filter-menu-actions">
            <button type="button" className="filter-clear" onClick={() => filters.forEach((filter) => filter.onChange(""))} disabled={activeCount === 0}>ล้างฟิลเตอร์</button>
            <button type="button" className="filter-done" onClick={() => setIsOpen(false)}><Check size={17} aria-hidden="true" /> ยืนยัน</button>
          </div>
        </div>
      )}
    </div>
  );
}
