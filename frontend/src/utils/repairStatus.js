const repairStatusLabels = {
  "รอซ่อม": "รอแอดมินมอบหมาย",
  "รอแอดมินจัดสรร": "รอแอดมินมอบหมาย",
  "รอช่างรับงาน": "รอช่างรับงาน",
  "ส่งคืนแอดมิน": "ส่งคืนแอดมิน",
  "รออนุมัติ": "รออนุมัติงบประมาณ",
  "ปิดงานก่อนครบขั้นตอน": "ปิดงานก่อนครบขั้นตอน",
  "รอรับงาน": "รอช่างรับงาน",
  "รับงาน": "รับงานแล้ว",
  "รับงานแล้ว": "รับงานแล้ว",
  "กำลังซ่อม": "กำลังดำเนินการซ่อม",
  "กำลังดำเนินการซ่อม": "กำลังดำเนินการซ่อม",
  "เสร็จเรียบร้อย": "ดำเนินเสร็จสิ้น",
  "เสร็จสิ้น": "ดำเนินเสร็จสิ้น",
  "ดำเนินเสร็จสิ้น": "ดำเนินเสร็จสิ้น",
};

export const getRepairStatusLabel = (status) =>
  repairStatusLabels[status] || status || "ไม่ระบุสถานะ";

const repairStatusClasses = {
  "รอซ่อม": "status-waiting-admin",
  "รอแอดมินจัดสรร": "status-waiting-admin",
  "รอช่างรับงาน": "status-waiting-technician",
  "รอรับงาน": "status-waiting-technician",
  "ส่งคืนแอดมิน": "status-returned-admin",
  "รออนุมัติ": "status-awaiting-approval",
  "รับงาน": "status-pending",
  "รับงานแล้ว": "status-pending",
  "กำลังซ่อม": "status-progress",
  "กำลังดำเนินการซ่อม": "status-progress",
  "เสร็จเรียบร้อย": "status-completed",
  "เสร็จสิ้น": "status-completed",
  "ดำเนินเสร็จสิ้น": "status-completed",
  "ซ่อมไม่ได้": "status-failed",
  "ส่งซ่อมภายนอก": "status-outsourced",
  "ส่งซ่อมภายนอก / คณะ": "status-outsourced",
  "ส่งซ่อมให้คณะ": "status-outsourced",
  "ปิดงานก่อนครบขั้นตอน": "status-closed-early",
};

export const getRepairStatusClass = (status) =>
  repairStatusClasses[status] || "status-default";
