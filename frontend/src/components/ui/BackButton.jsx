import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router-dom";

export default function BackButton({ to, label = "ย้อนกลับ", className = "" }) {
  const navigate = useNavigate();
  return (
    <button type="button" className={`system-back-button ${className}`.trim()} onClick={() => navigate(to)}>
      <ArrowLeft size={17} aria-hidden="true" />
      {label}
    </button>
  );
}
