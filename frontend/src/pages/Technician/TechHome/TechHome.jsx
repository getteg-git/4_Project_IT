import React, { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import AppBreadcrumb from "../../../components/AppBreadcrumb";
import RepairImageGallery from "../../../components/RepairImageGallery";
import {
  Banknote,
  Camera,
  CalendarDays,
  CircleCheck,
  CircleX,
  ClipboardList,
  Clock3,
  Eye,
  FileText,
  Image as ImageIcon,
  LogOut,
  Save,
  Trash2,
  UserRoundPlus,
  Wrench,
  X,
  Clock
} from "lucide-react";
import useToast from "../../../hooks/useToast";
import SearchField from "../../../components/ui/SearchField";
import Pagination from "../../../components/ui/Pagination";
import { getPageItems, REPAIR_PAGE_SIZE } from "../../../components/ui/paginationUtils";
import RepairCardNotes from "../../../components/RepairCardNotes";
import FilterBar from "../../../components/ui/FilterBar";
import RepairTimelineModal from "../../User/Timeline/RepairTimelineModal";
import "./TechHome.css";

function TechHome() {
  const navigate = useNavigate();
  const { toast, confirm } = useToast();

  const [currentUser, setCurrentUser] = useState(null);
  const [repairs, setRepairs] = useState([]);
  const [technicians, setTechnicians] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [jobView, setJobView] = useState("active");
  const [problemTypeFilter, setProblemTypeFilter] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false);
  const [assignmentRepair, setAssignmentRepair] = useState(null);
  const [isAssignOpen, setIsAssignOpen] = useState(false);
  const [chosenTechId, setChosenTechId] = useState("");
  const [isAssignSubmitting, setIsAssignSubmitting] = useState(false);

  // States สำหรับ Modal ปิดงาน
  const [modalOpen, setModalOpen] = useState(false);
  const [currentTicket, setCurrentTicket] = useState(null);
  const [closeType, setCloseType] = useState("");
  const [techDetail, setTechDetail] = useState("");
  const [actualCost, setActualCost] = useState("");
  const [completionImages, setCompletionImages] = useState([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const imagePickerRef = useRef(null);
  const cameraPickerRef = useRef(null);
  const completionImageUrlsRef = useRef(new Set());

  // States สำหรับ Modal ประเมินราคา
  const [estimateModalOpen, setEstimateModalOpen] = useState(false);
  const [estimateTicketId, setEstimateTicketId] = useState(null);
  const [estimatedCost, setEstimatedCost] = useState("");
  const [isStartingRepair, setIsStartingRepair] = useState(false);
  const [estimateBasePrice, setEstimateBasePrice] = useState(null);
  const [priceCheckStatus, setPriceCheckStatus] = useState("idle");

  useEffect(() => () => {
    completionImageUrlsRef.current.forEach((url) => URL.revokeObjectURL(url));
    completionImageUrlsRef.current.clear();
  }, []);

  const fetchMyJobs = async (user) => {
    try {
      setIsLoading(true);

      const [response, usersResponse] = await Promise.all([
        fetch("https://4projectit-production.up.railway.app/api/repairs"),
        fetch("https://4projectit-production.up.railway.app/api/users"),
      ]);

      if (response.ok) {
        const allRepairs = await response.json();

        // ช่างจะเห็นแค่งานที่ถูก Assign มาให้ตัวเองเท่านั้น
        const myJobs = allRepairs.filter(
          (r) => r.technician_id === user.id
        );

        setRepairs(myJobs);
      }
      if (usersResponse.ok) {
        const users = await usersResponse.json();
        setTechnicians(users.filter((item) => item.role === "technician"));
      }
    } catch (error) {
      console.error("ดึงข้อมูลงานซ่อมไม่สำเร็จ:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const storedUser = localStorage.getItem("user");

    if (storedUser) {
      const parsedUser = JSON.parse(storedUser);
      setCurrentUser(parsedUser);
      fetchMyJobs(parsedUser);
    } else {
      navigate("/");
    }
  }, [navigate]);

  const normalizedSearch = searchTerm
    .replace(/\s+/g, "")
    .replace(/วิทย์/g, "วิทยาศาสตร์")
    .toLowerCase();

  const problemTypeOptions = [
    ...new Set(
      repairs
        .map(
          (repair) =>
            repair.other_problem_type || repair.problem_type
        )
        .filter(Boolean)
    ),
  ].sort((a, b) => String(a).localeCompare(String(b), "th"));

  const filteredRepairs = repairs.filter((repair) => {
    if (
      problemTypeFilter &&
      (repair.other_problem_type || repair.problem_type) !==
      problemTypeFilter
    ) {
      return false;
    }

    if (!normalizedSearch) return true;

    const loc = (repair.location_name || "")
      .replace(/\s+/g, "")
      .toLowerCase();

    const floor = (repair.floor_name || "")
      .replace(/\s+/g, "")
      .toLowerCase();

    const room = (repair.room_number || "")
      .replace(/\s+/g, "")
      .toLowerCase();

    const equip = (repair.equipment_name || "")
      .replace(/\s+/g, "")
      .toLowerCase();

    const status = (repair.status || "")
      .replace(/\s+/g, "")
      .toLowerCase();

    const ticketId = (
      repair.ticket_number || String(repair.id)
    ).toLowerCase();

    return (
      loc.includes(normalizedSearch) ||
      floor.includes(normalizedSearch) ||
      room.includes(normalizedSearch) ||
      equip.includes(normalizedSearch) ||
      status.includes(normalizedSearch) ||
      ticketId.includes(normalizedSearch)
    );
  });

  const getStatusClass = (status) => {
    switch (status) {
      case "รอซ่อม":
        return "status-pending";

      case "รับงานแล้ว":
        return "status-pending";

      case "กำลังซ่อม":
        return "status-progress";

      case "เสร็จเรียบร้อย":
        return "status-completed";

      case "ซ่อมไม่ได้":
        return "status-failed";

      default:
        return "status-default";
    }
  };

  const formatDate = (dateString) => {
    const options = {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    };

    return new Date(dateString).toLocaleDateString(
      "th-TH",
      options
    );
  };

  const formatRoom = (roomNumber) => {
    if (!roomNumber) return "";

    const roomText = String(roomNumber).trim();

    return roomText.startsWith("ห้อง")
      ? `(${roomText})`
      : `(ห้อง ${roomText})`;
  };

  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const getAssignableTechnicians = (repair) =>
    technicians.filter((technician) =>
      technician.id !== currentUser?.id &&
      !technician.is_head_technician &&
      String(technician.department_id) === String(currentUser?.department_id) &&
      (technician.specialties || []).some(
        (specialtyId) => String(specialtyId) === String(repair.problem_type_id)
      )
    );

  const openAssignModal = (repair) => {
    setAssignmentRepair(repair);
    setChosenTechId("");
    setIsAssignOpen(true);
  };

  const handleAssignSubmit = async (event) => {
    event.preventDefault();
    if (!chosenTechId || !assignmentRepair) {
      toast.warning("กรุณาเลือกช่าง", {
        description: "เลือกช่างในสาขาที่มีความถนัดตรงกับประเภทงานก่อนมอบหมาย",
      });
      return;
    }

    setIsAssignSubmitting(true);
    try {
      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${assignmentRepair.id}/assign`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            technician_id: Number(chosenTechId),
            head_technician_id: currentUser.id,
          }),
        }
      );

      if (response.ok) {
        toast.success("มอบหมายงานสำเร็จ", {
          description: "ส่งต่องานให้ช่างในสาขาที่มีความถนัดตรงกันแล้ว",
        });
        setIsAssignOpen(false);
        setAssignmentRepair(null);
        fetchMyJobs(currentUser);
      } else {
        const data = await response.json();
        toast.error("มอบหมายงานไม่สำเร็จ", {
          description: data.error || "กรุณาลองใหม่อีกครั้ง",
        });
      }
    } catch (error) {
      console.error("Assign Job Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsAssignSubmitting(false);
    }
  };

  // รับงานไว้ก่อน โดยยังไม่เริ่มจับเวลาซ่อม
  const handleAcceptJob = async (repair) => {
    const approved = await confirm({
      title: "ยืนยันรับงานซ่อม",
      description: `คุณแน่ใจหรือไม่ว่าต้องการรับงาน ${
        repair.ticket_number || `#${repair.id}`
      }? งานจะอยู่ในสถานะรับงานแล้วจนกว่าคุณจะถึงหน้างานและเริ่มซ่อม`,
      confirmLabel: "ยืนยันรับงาน",
      cancelLabel: "ยกเลิก",
    });

    if (!approved) return;

    try {
      const formData = new FormData();

      formData.append("status", "รับงานแล้ว");
      formData.append("technician_id", currentUser.id);

      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${repair.id}/status`,
        {
          method: "PUT",
          body: formData,
        }
      );

      if (response.ok) {
        toast.success("รับงานสำเร็จ", {
          description: "งานจะอยู่ในรายการรอเข้าหน้างาน จนกว่าคุณจะกดเริ่มดำเนินการซ่อม",
        });

        fetchMyJobs(currentUser);
      } else {
        toast.error("รับงานไม่สำเร็จ", {
          description: "กรุณาลองใหม่อีกครั้ง",
        });
      }
    } catch (error) {
      console.error("Accept Job Error:", error);

      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    }
  };

  const handleStartRepair = async (repair) => {
    let basePrice = repair.base_price == null
      ? null
      : Number(repair.base_price);

    if (basePrice !== null && !Number.isFinite(basePrice)) {
      toast.error("โหลดเพดานราคาประเมินไม่สำเร็จ", {
        description: "ราคาทุนครุภัณฑ์ไม่ถูกต้อง กรุณาติดต่อผู้ดูแลระบบ",
      });
      return;
    }

    setEstimateTicketId(repair.id);
    setEstimatedCost("");
    clearCompletionImages();
    setEstimateBasePrice(basePrice);
    setPriceCheckStatus("idle");
    setIsStartingRepair(true);
    setEstimateModalOpen(true);
  };

  const handleReject = async (id) => {
    const confirmReject = await confirm({
      title: "ปฏิเสธงานซ่อม",
      description:
        "งานซ่อมนี้จะถูกส่งคืนให้ผู้ดูแลระบบพิจารณามอบหมายใหม่",
      confirmLabel: "ปฏิเสธงาน",
      variant: "danger",
    });

    if (!confirmReject) return;

    try {
      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${id}/reject`,
        {
          method: "PUT",
        }
      );

      if (response.ok) {
        toast.success("ส่งงานกลับผู้ดูแลแล้ว", {
          description: "ผู้ดูแลระบบจะพิจารณามอบหมายงานอีกครั้ง",
        });

        fetchMyJobs(currentUser);
      } else {
        toast.error("ปฏิเสธงานไม่สำเร็จ", {
          description: "กรุณาลองใหม่อีกครั้ง",
        });
      }
    } catch (error) {
      console.error("Reject Error:", error);
    }
  };

  // เปิดหน้าต่างประเมินราคา
  const openEstimateModal = (id) => {
    setEstimateTicketId(id);
    setEstimatedCost("");
    setIsStartingRepair(false);
    setEstimateBasePrice(null);
    setEstimateModalOpen(true);
  };

  const submitEstimate = async () => {
    setIsSubmitting(true);

    const costValue =
      estimatedCost === "" ? 0 : Number(estimatedCost);
    const useMultipart = isStartingRepair && completionImages.length > 0;
    const requestBody = useMultipart
      ? new FormData()
      : JSON.stringify({
          estimated_cost: costValue,
          ...(isStartingRepair && {
            tech_id: currentUser.id,
            force_admin_review: priceCheckStatus === "exceeded",
          }),
        });

    if (useMultipart) {
      requestBody.append("estimated_cost", String(costValue));
      requestBody.append("tech_id", String(currentUser.id));
      if (priceCheckStatus === "exceeded") {
        requestBody.append("force_admin_review", "true");
      }
      completionImages.forEach(({ file }) => requestBody.append("image", file));
    }

    try {
      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${estimateTicketId}/estimate`,
        {
          method: "PUT",
          ...(!useMultipart && {
            headers: { "Content-Type": "application/json" },
          }),
          body: requestBody,
        }
      );

      if (response.ok) {
        const data = await response.json();

        if (isStartingRepair && data.status === "รอซ่อม") {
          const ticket = repairs.find((repair) => repair.id === estimateTicketId);
          toast.warning("ส่งงานให้ผู้ดูแลระบบพิจารณาแล้ว", {
            description: `ราคาประเมินเกินเกณฑ์จุดคุ้มทุน (Ticket ${ticket?.ticket_number || `#${estimateTicketId}`})`,
          });
        } else if (isStartingRepair) {
          toast.success("เริ่มดำเนินการซ่อมแล้ว", {
            description: data.message || "เปลี่ยนสถานะงานเป็นกำลังซ่อมและเริ่มบันทึกเวลาแล้ว",
          });
        } else if (data.status === "รอซ่อม") {
          toast.info("งานรอการอนุมัติ", {
            description: data.message || "ระบบส่งเรื่องกลับให้ผู้ดูแลพิจารณาแล้ว",
          });
        } else {
          toast.success("ประเมินราคาสำเร็จ", {
            description: "คุณสามารถดำเนินการซ่อมต่อได้เลย",
          });
        }

        setEstimateModalOpen(false);
        setIsStartingRepair(false);
        setEstimateBasePrice(null);
        clearCompletionImages();
        fetchMyJobs(currentUser);
      } else {
        const responseText = await response.text();
        let errorMessage = "กรุณาลองใหม่อีกครั้ง";
        if (responseText) {
          try {
            const errorData = JSON.parse(responseText);
            errorMessage = errorData.error || errorData.message || errorMessage;
          } catch {
            errorMessage = responseText;
          }
        }

        toast.error("ส่งราคาประเมินไม่สำเร็จ", {
          description: errorMessage,
        });
      }
    } catch (error) {
      console.error("Estimate Error:", error);

      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้", {
        description: "ส่งราคาประเมินไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSubmitEstimate = async (event) => {
    event.preventDefault();
    if (isStartingRepair && priceCheckStatus !== "approved") return;
    await submitEstimate();
  };

  const handleCheckBreakEven = async () => {
    const cost = Number(estimatedCost);
    if (estimatedCost === "" || !Number.isFinite(cost) || cost < 0) {
      toast.error("กรุณากรอกราคาประเมินให้ถูกต้อง");
      return;
    }
    setPriceCheckStatus("checking");
    try {
      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${estimateTicketId}/estimate`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            estimated_cost: cost,
            tech_id: currentUser.id,
            check_only: true,
          }),
        }
      );
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "ตรวจสอบจุดคุ้มทุนไม่สำเร็จ");
      }
      const exceedsSingleRepairLimit =
        Number.isFinite(estimateBasePrice) &&
        cost > estimateBasePrice * 0.5;
      setPriceCheckStatus(
        data.requires_approval || exceedsSingleRepairLimit
          ? "exceeded"
          : "approved"
      );
    } catch (error) {
      console.error("Break-even Check Error:", error);
      setPriceCheckStatus("idle");
      toast.error("ตรวจสอบจุดคุ้มทุนไม่สำเร็จ", {
        description: error.message || "กรุณาลองใหม่อีกครั้ง",
      });
    }
  };

  // ปิดงานซ่อม
  const clearCompletionImages = () => {
    completionImages.forEach(({ previewUrl }) => {
      URL.revokeObjectURL(previewUrl);
      completionImageUrlsRef.current.delete(previewUrl);
    });
    setCompletionImages([]);
    if (imagePickerRef.current) imagePickerRef.current.value = "";
    if (cameraPickerRef.current) cameraPickerRef.current.value = "";
  };

  const openCloseModal = (id, type) => {
    clearCompletionImages();
    setCurrentTicket(id);
    setCloseType(type);
    setTechDetail("");
    setActualCost("");
    setModalOpen(true);
  };

  const handleCompletionImageSelection = (event) => {
    const selectedFiles = Array.from(event.target.files || []);
    event.target.value = "";

    const validFiles = selectedFiles.filter((file) => {
      if (!file.type.startsWith("image/")) {
        toast.warning("เลือกได้เฉพาะไฟล์รูปภาพ");
        return false;
      }
      if (file.size > 10 * 1024 * 1024) {
        toast.warning("ไฟล์รูปภาพมีขนาดใหญ่เกินไป", {
          description: `ไฟล์ "${file.name}" ต้องมีขนาดไม่เกิน 10 MB`,
        });
        return false;
      }
      return true;
    });
    const uniqueFiles = validFiles.filter((file) =>
      !completionImages.some((image) =>
        image.file.name === file.name &&
        image.file.size === file.size &&
        image.file.lastModified === file.lastModified
      )
    );
    const remainingSlots = 10 - completionImages.length;
    const filesToAdd = uniqueFiles.slice(0, remainingSlots);

    if (uniqueFiles.length > remainingSlots) {
      toast.warning("แนบรูปได้สูงสุด 10 รูป");
    }

    if (filesToAdd.length > 0) {
      const imagesToAdd = filesToAdd.map((file) => {
        const previewUrl = URL.createObjectURL(file);
        completionImageUrlsRef.current.add(previewUrl);
        return { file, previewUrl };
      });
      setCompletionImages((current) => [...current, ...imagesToAdd]);
      if (isStartingRepair) setPriceCheckStatus("idle");
    }
  };

  const removeCompletionImage = (index) => {
    const image = completionImages[index];
    if (!image) return;
    URL.revokeObjectURL(image.previewUrl);
    completionImageUrlsRef.current.delete(image.previewUrl);
    setCompletionImages((current) => current.filter((_, itemIndex) => itemIndex !== index));
    if (isStartingRepair) setPriceCheckStatus("idle");
  };

  const handleSubmitClose = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);

    const newStatus =
      closeType === "success"
        ? "เสร็จเรียบร้อย"
        : "ซ่อมไม่ได้";

    const costValue =
      closeType === "success" && actualCost
        ? actualCost
        : "0";

    const formData = new FormData();

    formData.append("status", newStatus);
    formData.append("technician_note", techDetail);
    formData.append("actual_cost", costValue);
    formData.append("technician_id", currentUser.id);
    completionImages.forEach(({ file }) => formData.append("image", file));

    try {
      const response = await fetch(
        `https://4projectit-production.up.railway.app/api/repairs/${currentTicket}/status`,
        {
          method: "PUT",
          body: formData,
        }
      );

      if (response.ok) {
        toast.success("บันทึกสถานะสำเร็จ", {
          description: `อัปเดตงานเป็นสถานะ “${newStatus}” แล้ว`,
        });

        setModalOpen(false);
        clearCompletionImages();
        fetchMyJobs(currentUser);
      } else {
        const errorData = await response.json();

        toast.error("บันทึกสถานะไม่สำเร็จ", {
          description:
            errorData.error || "กรุณาลองใหม่อีกครั้ง",
        });
      }
    } catch (error) {
      console.error("Close Job Error:", error);

      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleLogout = async () => {
    const approved = await confirm({
      title: "ยืนยันออกจากระบบ",
      description:
        "คุณแน่ใจหรือไม่ว่าต้องการออกจากระบบช่างเทคนิค?",
      confirmLabel: "ออกจากระบบ",
      cancelLabel: "ยกเลิก",
      variant: "danger",
    });

    if (approved) {
      localStorage.removeItem("user");
      navigate("/");
    }
  };

  const activeRepairs = filteredRepairs.filter(
    (repair) =>
      repair.status === "รอซ่อม" ||
      repair.status === "รับงานแล้ว"
  );

  const progressRepairs = filteredRepairs.filter(
    (repair) => repair.status === "กำลังซ่อม"
  );

  const completedRepairs = filteredRepairs.filter(
    (repair) =>
      repair.status !== "รอซ่อม" &&
      repair.status !== "รับงานแล้ว" &&
      repair.status !== "กำลังซ่อม"
  );

  const displayedRepairs =
    jobView === "active"
      ? activeRepairs
      : jobView === "progress"
        ? progressRepairs
        : completedRepairs;

  const paginatedRepairs = getPageItems(
    displayedRepairs,
    currentPage,
    REPAIR_PAGE_SIZE
  );

  const renderRepairCard = (repair) => (
    <div className="repair-item-card" key={repair.id}>
      <div className="card-top">
        <span className="repair-ticket-id">
          {repair.ticket_number ||
            `Ticket #${repair.id}`}
        </span>

        <span
          className={`status-badge ${getStatusClass(
            repair.status
          )}`}
        >
          {repair.status}
        </span>
      </div>

      <h3 className="repair-title">
        [{repair.problem_type}] {repair.location_name}
      </h3>

      <div className="repair-details-text">
        <p className="repair-info">
          <strong>ชั้น / พิกัด:</strong>{" "}
          {repair.floor_name || "-"}{" "}
          {formatRoom(repair.room_number)}
        </p>

        {repair.equipment_name && (
          <p className="repair-info equipment-info">
            <strong>
              <Wrench size={16} aria-hidden="true" />{" "}
              อุปกรณ์:
            </strong>{" "}
            {repair.equipment_name}
          </p>
        )}

        <p className="repair-info">
          <strong>อาการที่แจ้ง:</strong>{" "}
          {repair.description}
        </p>

        <RepairCardNotes repair={repair} />
      </div>

      <div className="card-actions-row">
        <span className="repair-date">
          <CalendarDays size={16} aria-hidden="true" />{" "}
          แจ้งเมื่อ: {formatDate(repair.created_at)}
        </span>

        <div
          className="action-buttons"
          style={{
            display: "flex",
            gap: "8px",
          }}
        >
          <button
            className="btn-details"
            onClick={() => openDetailsModal(repair)}
          >
            <Eye size={16} aria-hidden="true" />{" "}
            รายละเอียด
          </button>

          <button
            className="btn-timeline"
            onClick={() => {
              setSelectedRepair(repair);
              setIsTimelineOpen(true);
            }}
            style={{
              backgroundColor: "#f1f5f9",
              color: "#334155",
              border: "none",
              padding: "6px 12px",
              borderRadius: "6px",
              cursor: "pointer",
              display: "flex",
              alignItems: "center",
              gap: "4px",
              fontSize: "0.85rem",
            }}
          >
            <Clock size={16} aria-hidden="true" />{" "}
            ไทม์ไลน์
          </button>

          {currentUser?.is_head_technician && repair.status === "รอซ่อม" && (
            <button
              className="btn-details"
              type="button"
              onClick={() => openAssignModal(repair)}
            >
              <UserRoundPlus size={16} aria-hidden="true" /> มอบหมายงาน
            </button>
          )}
        </div>
      </div>

      {(repair.status === "รอซ่อม" ||
        repair.status === "รับงานแล้ว") && (
        <div className="job-decision-row">
          {repair.status === "รับงานแล้ว" ? (
            <button
              className="btn-accept"
              onClick={() => handleStartRepair(repair)}
            >
              <CircleCheck
                size={17}
                aria-hidden="true"
              />{" "}
              เริ่มดำเนินการซ่อม
            </button>
          ) : repair.admin_note &&
            repair.admin_note.trim() !== "" ? (
            <button
              className="btn-waiting-approval"
              disabled
            >
              <Clock3
                size={17}
                aria-hidden="true"
              />{" "}
              รอแอดมินอนุมัติงบ
            </button>
          ) : (
            <button
              className="btn-accept"
              onClick={() => handleAcceptJob(repair)}
            >
              <CircleCheck
                size={17}
                aria-hidden="true"
              />{" "}
              รับงาน
            </button>
          )}

          {repair.status === "รอซ่อม" && (
            <button
              className="btn-reject"
              onClick={() =>
                handleReject(repair.id)
              }
            >
              <X size={17} aria-hidden="true" />{" "}
              ปฏิเสธงาน
            </button>
          )}
        </div>
      )}

      {repair.status === "กำลังซ่อม" && (
        <div
          className="job-completion-row"
          style={{ flexWrap: "wrap" }}
        >
          {(repair.estimated_cost === null ||
            repair.estimated_cost === undefined) && (
            <button
              className="btn-estimate"
              onClick={() =>
                openEstimateModal(repair.id)
              }
              style={{
                backgroundColor: "#f59e0b",
                color: "#fff",
                border: "none",
                padding: "8px 16px",
                borderRadius: "6px",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                gap: "6px",
                fontWeight: "bold",
              }}
            >
              <Banknote
                size={17}
                aria-hidden="true"
              />{" "}
              ประเมินราคา
            </button>
          )}

          <button
            className="btn-complete"
            onClick={() =>
              openCloseModal(
                repair.id,
                "success"
              )
            }
          >
            <CircleCheck
              size={17}
              aria-hidden="true"
            />{" "}
            ซ่อมสำเร็จ
          </button>

          <button
            className="btn-cannot-fix"
            onClick={() =>
              openCloseModal(
                repair.id,
                "fail"
              )
            }
          >
            <CircleX
              size={17}
              aria-hidden="true"
            />{" "}
            ซ่อมไม่สำเร็จ
          </button>
        </div>
      )}
    </div>
  );

  return (
    <div className={`tech-container ${!isLoading && displayedRepairs.length === 0 ? "has-empty-jobs" : ""}`}>
      <div className="tech-wrapper">

        <AppBreadcrumb items={[
          { label: "พื้นที่ช่าง" },
          { label: "งานซ่อมที่ได้รับมอบหมาย" },
        ]} />

        <div className="tech-header page-header">
          <div>
            <h2>
              <ClipboardList
                size={24}
                aria-hidden="true"
              />{" "}
              งานซ่อมที่ได้รับมอบหมาย
            </h2>

            <p>
              ยินดีต้อนรับคุณ{" "}
              {currentUser
                ? currentUser.full_name ||
                currentUser.username
                : "กำลังโหลดข้อมูลผู้ใช้..."}
              {" — ติดตามและอัปเดตสถานะงานได้ที่นี่"}
            </p>
          </div>

          <div className="header-actions page-logout-group">
            <button
              className="btn-logout page-logout"
              onClick={handleLogout}
              aria-label="ออกจากระบบ"
              title="ออกจากระบบ"
            >
              <LogOut
                size={17}
                aria-hidden="true"
                className="logout-icon"
              />
              <span className="logout-label">ออกจากระบบ</span>
            </button>
          </div>
        </div>

        <div className="tech-workspace">
        <div className="search-section">
          <SearchField
            id="technician-repair-search"
            label="ค้นหางานที่รับผิดชอบ"
            placeholder="หมายเลข Ticket สถานที่ ห้อง อุปกรณ์ หรือสถานะ"
            value={searchTerm}
            onChange={(value) => {
              setSearchTerm(value);
              setCurrentPage(1);
            }}
          />

          <FilterBar
            filters={[
              {
                id: "problem-type",
                label: "ประเภทงาน",
                value: problemTypeFilter,
                options: problemTypeOptions,
                onChange: (value) => {
                  setProblemTypeFilter(value);
                  setCurrentPage(1);
                },
              },
            ]}
          />
        </div>

        {isLoading ? (
          <div className="loading-state">
            กำลังโหลดรายการงานของคุณ...
          </div>
        ) : (
          <div className={`job-view-area ${displayedRepairs.length === 0 ? "is-empty" : ""}`}>
            <div
              className="job-view-tabs"
              role="tablist"
              aria-label="เลือกประเภทงาน"
            >
              <button
                type="button"
                role="tab"
                aria-selected={
                  jobView === "active"
                }
                className={
                  jobView === "active"
                    ? "is-active"
                    : ""
                }
                onClick={() => {
                  setJobView("active");
                  setCurrentPage(1);
                }}
              >
                รอรับงาน / ไปหน้างาน{" "}
                <span>{activeRepairs.length}</span>
              </button>

              <button
                type="button"
                role="tab"
                aria-selected={
                  jobView === "progress"
                }
                className={
                  jobView === "progress"
                    ? "is-active"
                    : ""
                }
                onClick={() => {
                  setJobView("progress");
                  setCurrentPage(1);
                }}
              >
                กำลังซ่อม{" "}
                <span>{progressRepairs.length}</span>
              </button>

              <button
                type="button"
                role="tab"
                aria-selected={
                  jobView === "completed"
                }
                className={
                  jobView === "completed"
                    ? "is-active"
                    : ""
                }
                onClick={() => {
                  setJobView("completed");
                  setCurrentPage(1);
                }}
              >
                เสร็จแล้ว{" "}
                <span>
                  {completedRepairs.length}
                </span>
              </button>
            </div>

            {displayedRepairs.length > 0 ? (
              <>
                <div className="repair-list">
                  {paginatedRepairs.map(
                    renderRepairCard
                  )}
                </div>

                <Pagination
                  currentPage={currentPage}
                  totalItems={displayedRepairs.length}
                  onPageChange={setCurrentPage}
                  pageSize={REPAIR_PAGE_SIZE}
                />
              </>
            ) : (
              <div className="tech-empty-state">
                <p>
                  {jobView === "active"
                    ? "ไม่มีงานที่รอรับหรือรอเข้าหน้างาน"
                    : jobView === "progress"
                      ? "ไม่มีงานที่กำลังซ่อม"
                      : "ยังไม่มีประวัติงานที่เสร็จแล้ว"}
                </p>
              </div>
            )}
          </div>
        )}
        </div>
      </div>

      {/* POPUP 1: ดูรายละเอียดงาน */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <button
              className="close-btn"
              type="button"
              onClick={() =>
                setIsDetailsOpen(false)
              }
              aria-label="ปิดรายละเอียด"
            >
              <X aria-hidden="true" />
            </button>

            <div className="modal-header-center">
              <h3 className="modal-title-primary">
                <FileText
                  size={22}
                  aria-hidden="true"
                />{" "}
                ข้อมูลการแจ้งซ่อมอย่างละเอียด
              </h3>

              <p className="modal-subtitle">
                Ticket:{" "}
                {selectedRepair.ticket_number ||
                  `#${selectedRepair.id}`}
              </p>
            </div>

            <div className="modal-details-content">
              <RepairImageGallery
                images={selectedRepair.images}
                ticketLabel={selectedRepair.ticket_number || `#${selectedRepair.id}`}
              />

              <div className="info-list-container">
                <p className="repair-info-item">
                  <strong>
                    สถานะปัจจุบัน:
                  </strong>{" "}
                  <span
                    className={`status-badge ${getStatusClass(
                      selectedRepair.status
                    )}`}
                  >
                    {selectedRepair.status}
                  </span>
                </p>

                <p className="repair-info-item">
                  <strong>หมวดหมู่งาน:</strong>{" "}
                  {selectedRepair.problem_type}
                </p>

                <p className="repair-info-item">
                  <strong>สถานที่:</strong>{" "}
                  {selectedRepair.location_name}
                </p>

                <p className="repair-info-item">
                  <strong>ชั้น / พิกัด:</strong>{" "}
                  {selectedRepair.floor_name ||
                    "-"}{" "}
                  {formatRoom(
                    selectedRepair.room_number
                  )}
                </p>

                {selectedRepair.equipment_name && (
                  <p className="repair-info-item">
                    <strong>
                      อุปกรณ์ชำรุด:
                    </strong>{" "}
                    {selectedRepair.equipment_name}
                  </p>
                )}

                <p className="repair-info-item">
                  <strong>อีเมลผู้แจ้ง:</strong>{" "}
                  {selectedRepair.reporter_email}
                </p>

                <p className="repair-info-item">
                  <strong>วันที่แจ้งเรื่อง:</strong>{" "}
                  {formatDate(
                    selectedRepair.created_at
                  )}
                </p>
              </div>

              <div className="issue-desc-box">
                <strong>
                  <FileText
                    size={17}
                    aria-hidden="true"
                  />{" "}
                  รายละเอียดปัญหา:
                </strong>

                <p className="issue-desc-text">
                  {selectedRepair.description}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {isAssignOpen && assignmentRepair && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button
              className="close-btn"
              type="button"
              disabled={isAssignSubmitting}
              onClick={() => {
                setIsAssignOpen(false);
                setAssignmentRepair(null);
              }}
              aria-label="ปิดหน้าต่าง"
            >
              <X aria-hidden="true" />
            </button>

            <h3 className="modal-title-primary">
              <UserRoundPlus size={22} aria-hidden="true" /> มอบหมายงานให้ช่าง
            </h3>
            <p className="modal-subtitle">
              {assignmentRepair.ticket_number || `#${assignmentRepair.id}`} ·{" "}
              {assignmentRepair.problem_type}
            </p>

            <form onSubmit={handleAssignSubmit}>
              <div className="input-group">
                <label htmlFor="head-technician-assignee">
                  ช่างผู้รับงาน <span className="required">*</span>
                </label>
                <select
                  id="head-technician-assignee"
                  value={chosenTechId}
                  onChange={(event) => setChosenTechId(event.target.value)}
                  required
                  disabled={getAssignableTechnicians(assignmentRepair).length === 0}
                >
                  <option value="">-- เลือกช่างที่ถนัดงานนี้ --</option>
                  {getAssignableTechnicians(assignmentRepair).map((technician) => (
                    <option key={technician.id} value={technician.id}>
                      {technician.full_name}
                    </option>
                  ))}
                </select>
                {getAssignableTechnicians(assignmentRepair).length === 0 && (
                  <small>ไม่พบช่างในสาขาที่มีความถนัดตรงกับประเภทงานนี้</small>
                )}
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="btn-cancel"
                  disabled={isAssignSubmitting}
                  onClick={() => {
                    setIsAssignOpen(false);
                    setAssignmentRepair(null);
                  }}
                >
                  ยกเลิก
                </button>
                <button
                  type="submit"
                  className="btn-submit"
                  disabled={isAssignSubmitting || getAssignableTechnicians(assignmentRepair).length === 0}
                >
                  {isAssignSubmitting ? "กำลังมอบหมาย..." : "ยืนยันมอบหมายงาน"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* POPUP 2: ประเมินราคา */}
      {estimateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button
              className="close-btn"
              type="button"
              disabled={isSubmitting}
              onClick={() => {
                setEstimateModalOpen(false);
                setIsStartingRepair(false);
                setEstimateBasePrice(null);
                clearCompletionImages();
              }}
              aria-label="ปิดหน้าต่าง"
            >
              <X aria-hidden="true" />
            </button>

            <h3 className="modal-title-primary">
              <Banknote
                size={23}
                aria-hidden="true"
              />{" "}
              {isStartingRepair ? "เริ่มซ่อมและประเมินราคา" : "ประเมินราคาก่อนดำเนินการซ่อม"}
            </h3>

            <p className="modal-subtitle">
              {isStartingRepair
                ? "ยืนยันเริ่มซ่อมเพื่อเปลี่ยนสถานะงานและเริ่มบันทึกเวลา"
                : "รหัสใบงาน: "}
              {!isStartingRepair && (
                <>
              {repairs.find(
                (r) => r.id === estimateTicketId
              )?.ticket_number ||
                `#${estimateTicketId}`}
                </>
              )}
            </p>

            <form onSubmit={handleSubmitEstimate}>
              {isStartingRepair && (
                <div className="input-group">
                  <label>
                    <Camera size={17} aria-hidden="true" /> รูปภาพก่อนเริ่มซ่อม
                  </label>
                  <p className="completion-image-hint">
                    แนบรูปได้ตามต้องการ สูงสุด 10 รูป (ไม่เกิน 10 MB ต่อรูป)
                  </p>
                  <div className="completion-image-actions">
                    <button
                      type="button"
                      className="completion-image-picker"
                      onClick={() => imagePickerRef.current?.click()}
                    >
                      <ImageIcon size={18} aria-hidden="true" /> เลือกหลายรูป
                    </button>
                    <button
                      type="button"
                      className="completion-camera-picker"
                      onClick={() => cameraPickerRef.current?.click()}
                    >
                      <Camera size={18} aria-hidden="true" /> ถ่ายรูป
                    </button>
                  </div>
                  <input
                    ref={imagePickerRef}
                    className="completion-image-input"
                    type="file"
                    accept="image/*"
                    multiple
                    onChange={handleCompletionImageSelection}
                    aria-label="เลือกรูปก่อนเริ่มซ่อม"
                  />
                  <input
                    ref={cameraPickerRef}
                    className="completion-image-input"
                    type="file"
                    accept="image/*"
                    capture="environment"
                    onChange={handleCompletionImageSelection}
                    aria-label="ถ่ายรูปก่อนเริ่มซ่อม"
                  />
                  {completionImages.length > 0 && (
                    <div className="completion-image-preview-grid">
                      {completionImages.map(({ file, previewUrl }, index) => (
                        <div className="completion-image-preview" key={`${file.name}-${file.lastModified}-${index}`}>
                          <img src={previewUrl} alt={`รูปก่อนเริ่มซ่อม ${index + 1}: ${file.name}`} />
                          <span title={file.name}>{file.name}</span>
                          <button
                            type="button"
                            onClick={() => removeCompletionImage(index)}
                            aria-label={`ลบรูป ${file.name}`}
                          >
                            <Trash2 size={16} aria-hidden="true" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                  <p className="completion-image-count" aria-live="polite">
                    เลือกแล้ว {completionImages.length}/10 รูป
                  </p>
                </div>
              )}

              {isStartingRepair && Number.isFinite(estimateBasePrice) && (
                <>
                  <div className="estimate-budget-note">
                    <div>
                      <span>ราคาทุนครุภัณฑ์</span>
                      <strong>{estimateBasePrice.toLocaleString()} บาท</strong>
                    </div>
                    <div>
                      <label htmlFor="break-even-point">จุดคุ้มทุนต่อรายการ (50%)</label>
                      <input
                        id="break-even-point"
                        type="text"
                        value={`${(estimateBasePrice * 0.5).toLocaleString()} บาท`}
                        readOnly
                        aria-describedby="break-even-help"
                      />
                    </div>
                  </div>
                  <p id="break-even-help" className="estimate-warning-note">
                    หากเกินจุดคุ้มทุน ระบบจะส่งรายการให้ผู้ดูแลระบบพิจารณา โดยตรวจสอบยอดสะสมร่วมด้วย
                  </p>
                </>
              )}
              <div className="input-group">
                <label>
                  <Banknote
                    size={17}
                    aria-hidden="true"
                  />{" "}
                  ราคาประเมินเบื้องต้น (บาท){" "}
                  <span className="required">
                    *
                  </span>
                </label>

                <input
                  type="number"
                  min="0"
                  step="0.01"
                  className="cost-input"
                  placeholder="ใส่ราคาประเมิน (กรอก 0 หากไม่เสียค่าใช้จ่าย)"
                  value={estimatedCost}
                  onChange={(e) => {
                    setEstimatedCost(e.target.value);
                    setPriceCheckStatus("idle");
                  }}
                  required
                />

                {isStartingRepair && estimatedCost !== "" && (
                  <p
                    className={`estimate-check-status estimate-check-${priceCheckStatus}`}
                    role="status"
                    aria-live="polite"
                  >
                    {priceCheckStatus === "checking" && "กำลังตรวจสอบราคา..."}
                    {priceCheckStatus === "approved" && "ผ่านเกณฑ์ — ยืนยันการซ่อมได้"}
                    {priceCheckStatus === "exceeded" &&
                      `เนื่องจากราคาประเมิน (${Number(estimatedCost).toLocaleString("th-TH")} บาท) สูงกว่าจุดคุ้มทุน ระบบจำเป็นต้องส่งรายการนี้ให้ "แอดมิน / ผู้ดูแลระบบ" พิจารณาอนุมัติก่อนเริ่มซ่อม`}
                  </p>
                )}

                <small className="cost-hint">
                  * กดตรวจสอบราคาก่อน ระบบจะประเมินยอดต่อรายการและตรวจยอดสะสมอีกครั้งที่เซิร์ฟเวอร์
                </small>
              </div>

              <div className="modal-actions">
                {isStartingRepair && priceCheckStatus !== "exceeded" && (
                  <button
                    type="button"
                    className="btn-cancel"
                    disabled={isSubmitting}
                    onClick={() => {
                      setEstimateModalOpen(false);
                      setIsStartingRepair(false);
                      setEstimateBasePrice(null);
                      clearCompletionImages();
                    }}
                  >
                    ยกเลิก
                  </button>
                )}
                {!isStartingRepair ? (
                  <button type="submit" className="btn-accept" disabled={isSubmitting}>
                    {isSubmitting ? "กำลังบันทึก..." : "บันทึกราคาประเมิน"}
                  </button>
                ) : priceCheckStatus === "approved" ? (
                  <button type="submit" className="btn-accept" disabled={isSubmitting}>
                    {isSubmitting ? "กำลังบันทึก..." : "ยืนยันเริ่มซ่อม"}
                  </button>
                ) : priceCheckStatus === "exceeded" && !isSubmitting ? (
                  <>
                    <button
                      type="button"
                      className="btn-cancel"
                      onClick={() => setPriceCheckStatus("idle")}
                    >
                      ยกเลิก / แก้ไขราคา
                    </button>
                    <button
                      type="button"
                      className="btn-accept"
                      disabled={isSubmitting}
                      onClick={submitEstimate}
                    >
                      ยืนยันส่งแอดมินพิจารณา
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    className="btn-accept"
                    disabled={
                      isSubmitting ||
                      estimatedCost === "" ||
                      priceCheckStatus === "checking"
                    }
                    onClick={handleCheckBreakEven}
                  >
                    {isSubmitting
                      ? "กำลังบันทึก..."
                      : priceCheckStatus === "checking"
                        ? "กำลังตรวจสอบราคา..."
                        : "ตรวจสอบราคา"}
                  </button>
                )}
              </div>
            </form>
          </div>
        </div>
      )}

      {/* POPUP 3: ปิดงาน */}
      {modalOpen && (
        <div className="modal-overlay">
          <div className="modal-box completion-modal-box">
            <button
              className="close-btn"
              type="button"
              onClick={() => {
                setModalOpen(false);
                clearCompletionImages();
              }}
              aria-label="ปิดหน้าต่าง"
            >
              <X aria-hidden="true" />
            </button>

            <h3
              className={
                closeType === "success"
                  ? "modal-title-success"
                  : "modal-title-fail"
              }
            >
              {closeType === "success" ? (
                <>
                  <CircleCheck
                    size={23}
                    aria-hidden="true"
                  />{" "}
                  ปิดงาน: ซ่อมเสร็จเรียบร้อย
                </>
              ) : (
                <>
                  <CircleX
                    size={23}
                    aria-hidden="true"
                  />{" "}
                  ปิดงาน: ซ่อมไม่สำเร็จ
                </>
              )}
            </h3>

            <p className="modal-subtitle">
              รหัสใบงาน:{" "}
              {repairs.find(
                (r) => r.id === currentTicket
              )?.ticket_number ||
                `#${currentTicket}`}
            </p>

            <form onSubmit={handleSubmitClose}>
              {closeType === "success" && (
                <div className="input-group">
                  <label>
                    <Banknote
                      size={17}
                      aria-hidden="true"
                    />{" "}
                    ค่าใช้จ่ายเบิกจริง (บาท){" "}
                    <span className="required">
                      *
                    </span>
                  </label>

                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="cost-input"
                    placeholder="ใส่ยอดใช้จ่ายจริงเพื่อเก็บเป็นสถิติสะสม"
                    value={actualCost}
                    onChange={(e) =>
                      setActualCost(
                        e.target.value
                      )
                    }
                    required
                  />
                </div>
              )}

              <div className="input-group">
                <label>
                  <FileText
                    size={17}
                    aria-hidden="true"
                  />{" "}
                  {closeType === "success"
                    ? "รายละเอียดการซ่อมแซม"
                    : "ระบุเหตุผลที่ขัดข้อง"}{" "}
                  <span className="required">
                    *
                  </span>
                </label>

                <textarea
                  rows="4"
                  className="textarea-input"
                  placeholder={
                    closeType === "success"
                      ? "เช่น ล้างแอร์ เติมน้ำยาแอร์..."
                      : "เช่น อะไหล่ต้องสั่งจากต่างประเทศ อาการหนัก..."
                  }
                  value={techDetail}
                  onChange={(e) =>
                    setTechDetail(e.target.value)
                  }
                  required
                ></textarea>
              </div>

              <div className="input-group">
                <label>
                  <Camera size={17} aria-hidden="true" /> รูปถ่ายหลังดำเนินการ
                </label>
                <p className="completion-image-hint">
                  แนบรูปได้ตามต้องการ สูงสุด 10 รูป (ไม่เกิน 10 MB ต่อรูป)
                </p>
                <div className="completion-image-actions">
                  <button
                    type="button"
                    className="completion-image-picker"
                    onClick={() => imagePickerRef.current?.click()}
                  >
                    <ImageIcon size={18} aria-hidden="true" /> เลือกหลายรูป
                  </button>
                  <button
                    type="button"
                    className="completion-camera-picker"
                    onClick={() => cameraPickerRef.current?.click()}
                  >
                    <Camera size={18} aria-hidden="true" /> ถ่ายรูป
                  </button>
                </div>
                <input
                  ref={imagePickerRef}
                  className="completion-image-input"
                  type="file"
                  accept="image/*"
                  multiple
                  onChange={handleCompletionImageSelection}
                  aria-label="เลือกรูปภาพหลายรูป"
                />
                <input
                  ref={cameraPickerRef}
                  className="completion-image-input"
                  type="file"
                  accept="image/*"
                  capture="environment"
                  onChange={handleCompletionImageSelection}
                  aria-label="ถ่ายรูปด้วยกล้อง"
                />
                {completionImages.length > 0 && (
                  <div className="completion-image-preview-grid">
                    {completionImages.map(({ file, previewUrl }, index) => (
                      <div className="completion-image-preview" key={`${file.name}-${file.lastModified}-${index}`}>
                        <img src={previewUrl} alt={`รูปประกอบ ${index + 1}: ${file.name}`} />
                        <span title={file.name}>{file.name}</span>
                        <button
                          type="button"
                          onClick={() => removeCompletionImage(index)}
                          aria-label={`ลบรูป ${file.name}`}
                        >
                          <Trash2 size={16} aria-hidden="true" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <p className="completion-image-count" aria-live="polite">
                  เลือกแล้ว {completionImages.length}/10 รูป
                </p>
              </div>

              <div className="modal-actions">
                <button
                  type="submit"
                  className={
                    closeType === "success"
                      ? "btn-submit-success"
                      : "btn-submit-fail"
                  }
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    "กำลังบันทึก..."
                  ) : (
                    <>
                      <Save
                        size={18}
                        aria-hidden="true"
                      />{" "}
                      ยืนยันปิดงาน
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* POPUP 4: Timeline */}
      <RepairTimelineModal
        repairId={selectedRepair?.id}
        images={selectedRepair?.images}
        ticketLabel={selectedRepair?.ticket_number || `#${selectedRepair?.id}`}
        isOpen={isTimelineOpen}
        onClose={() => {
          setIsTimelineOpen(false);
          setSelectedRepair(null);
        }}
      />
    </div>
  );
}

export default TechHome;