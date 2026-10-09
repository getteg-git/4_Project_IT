import { useEffect, useState } from "react";
import { Camera, ZoomIn, X } from "lucide-react";
import "./RepairImageGallery.css";

const API_BASE_URL = "https://4projectit-production.up.railway.app";

function RepairImageGallery({ images = [], ticketLabel = "งานซ่อม" }) {
  const [activeImage, setActiveImage] = useState(null);
  const repairImages = Array.isArray(images) ? images : [];

  useEffect(() => {
    if (!activeImage) return undefined;

    const handleKeyDown = (event) => {
      if (event.key === "Escape") setActiveImage(null);
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activeImage]);

  return (
    <>
      <div className="repair-image-gallery">
        {repairImages.length > 0 ? (
          repairImages.map((image, index) => {
            const imageType = image.type === "after" ? "หลังซ่อม" : "ก่อนซ่อม";
            const alt = `รูป${imageType}ของ ${ticketLabel}`;

            return (
              <figure className="repair-image-card" key={image.id || image.url || index}>
                <button
                  className="repair-image-trigger"
                  type="button"
                  onClick={() => setActiveImage({
                    src: `${API_BASE_URL}${image.url}`,
                    alt,
                  })}
                  aria-label={`ขยายรูป${imageType}ของ ${ticketLabel}`}
                >
                  <img
                    src={`${API_BASE_URL}${image.url}`}
                    alt={alt}
                    className="repair-image-thumbnail"
                  />
                  <span className="repair-image-zoom-hint">
                    <ZoomIn size={15} aria-hidden="true" /> ขยายรูป
                  </span>
                </button>
                <figcaption className="repair-image-caption">
                  <Camera size={14} aria-hidden="true" />
                  {imageType}
                </figcaption>
              </figure>
            );
          })
        ) : (
          <div className="repair-image-empty">ไม่มีรูปภาพประกอบ</div>
        )}
      </div>

      {activeImage && (
        <div
          className="repair-image-lightbox"
          role="presentation"
          onClick={() => setActiveImage(null)}
        >
          <button
            className="repair-image-lightbox-close"
            type="button"
            onClick={() => setActiveImage(null)}
            aria-label="ปิดรูปภาพ"
          >
            <X size={23} aria-hidden="true" />
          </button>
          <figure
            className="repair-image-lightbox-content"
            role="dialog"
            aria-modal="true"
            aria-label={activeImage.alt}
            onClick={(event) => event.stopPropagation()}
          >
            <img src={activeImage.src} alt={activeImage.alt} />
            <figcaption>
              {activeImage.alt} · กด Escape หรือคลิกพื้นหลังเพื่อปิด
            </figcaption>
          </figure>
        </div>
      )}
    </>
  );
}

export default RepairImageGallery;
