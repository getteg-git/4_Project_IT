import { useEffect, useState } from "react";
import { Camera, ZoomIn, X } from "lucide-react";
import { API_BASE_URL } from "../api/baseUrl";
import "./RepairImageGallery.css";

const EMPTY_IMAGES = [];

function RepairImageGallery({ images = EMPTY_IMAGES, ticketLabel = "งานซ่อม", requestHeaders }) {
  const [activeImage, setActiveImage] = useState(null);
  const [loadedImages, setLoadedImages] = useState({});
  const [failedImages, setFailedImages] = useState({});
  const repairImages = Array.isArray(images) ? images : EMPTY_IMAGES;
  const headersKey = JSON.stringify(requestHeaders || {});

  useEffect(() => {
    let active = true;
    const objectUrls = [];
    const headers = JSON.parse(headersKey);
    setLoadedImages({});
    setFailedImages({});

    repairImages.forEach(async (image, index) => {
      const key = String(image.id || image.url || index);
      const imageUrl = image.url.startsWith("http")
        ? image.url
        : `${API_BASE_URL}${image.url}`;
      try {
        const response = await fetch(imageUrl, {
          credentials: "include",
          headers,
        });
        if (!response.ok) throw new Error(`Image request failed: ${response.status}`);
        const objectUrl = URL.createObjectURL(await response.blob());
        objectUrls.push(objectUrl);
        if (active) {
          setLoadedImages((current) => ({ ...current, [key]: objectUrl }));
        } else {
          URL.revokeObjectURL(objectUrl);
        }
      } catch {
        if (active) setFailedImages((current) => ({ ...current, [key]: true }));
      }
    });

    return () => {
      active = false;
      objectUrls.forEach((url) => URL.revokeObjectURL(url));
    };
  }, [headersKey, repairImages]);

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
            const imageType = image.type === "after" ? "หลังดำเนินการ" : "ก่อนซ่อม";
            const alt = `รูป${imageType}ของ ${ticketLabel}`;
            const key = String(image.id || image.url || index);
            const imageSrc = loadedImages[key];

            return (
              <figure className="repair-image-card" key={key}>
                <button
                  className="repair-image-trigger"
                  type="button"
                  onClick={() => imageSrc && setActiveImage({ src: imageSrc, alt })}
                  disabled={!imageSrc}
                  aria-label={`ขยายรูป${imageType}ของ ${ticketLabel}`}
                >
                  {imageSrc ? (
                    <>
                      <img src={imageSrc} alt={alt} className="repair-image-thumbnail" />
                      <span className="repair-image-zoom-hint">
                        <ZoomIn size={15} aria-hidden="true" /> ขยายรูป
                      </span>
                    </>
                  ) : (
                    <span>{failedImages[key] ? "ไม่สามารถโหลดรูปภาพได้" : "กำลังโหลดรูปภาพ..."}</span>
                  )}
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
