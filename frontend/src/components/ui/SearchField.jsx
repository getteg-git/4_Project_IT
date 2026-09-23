import { Search, X } from "lucide-react";

export default function SearchField({ id, label = "ค้นหารายการ", placeholder, value, onChange }) {
  return (
    <div className="system-search">
      <div className="system-search-meta">
        <label htmlFor={id}>{label}</label>
      </div>
      <div className="system-search-field">
        <Search size={19} aria-hidden="true" />
        <input
          id={id}
          type="search"
          value={value}
          placeholder={placeholder}
          onChange={(event) => onChange(event.target.value)}
        />
        {value && (
          <button type="button" onClick={() => onChange("")} aria-label="ล้างคำค้นหา">
            <X size={17} aria-hidden="true" />
          </button>
        )}
      </div>
    </div>
  );
}
