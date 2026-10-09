import { ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";
import "./AppBreadcrumb.css";

function AppBreadcrumb({ items }) {
  return (
    <nav className="app-breadcrumb" aria-label="ตำแหน่งปัจจุบัน">
      {items.map((item, index) => {
        const isCurrent = index === items.length - 1;

        return (
          <span className="app-breadcrumb-item" key={`${item.label}-${index}`}>
            {index > 0 && <ChevronRight size={15} aria-hidden="true" />}
            {item.to && !isCurrent ? (
              <Link to={item.to}>{item.label}</Link>
            ) : (
              <span aria-current={isCurrent ? "page" : undefined}>{item.label}</span>
            )}
          </span>
        );
      })}
    </nav>
  );
}

export default AppBreadcrumb;
