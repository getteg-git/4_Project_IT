package repair

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestBuildBreakEvenSummary(t *testing.T) {
	tests := []struct {
		name            string
		hasEquipment    bool
		basePrice       float64
		accumulatedCost float64
		wantLevel       string
		wantPercent     float64
	}{
		{name: "below 30 percent", hasEquipment: true, basePrice: 1000, accumulatedCost: 299.99, wantLevel: "normal", wantPercent: 29.999},
		{name: "30 percent advisory", hasEquipment: true, basePrice: 1000, accumulatedCost: 300, wantLevel: "warning", wantPercent: 30},
		{name: "50 percent high alert", hasEquipment: true, basePrice: 1000, accumulatedCost: 500, wantLevel: "high", wantPercent: 50},
		{name: "no linked equipment", accumulatedCost: 500, wantLevel: "no_equipment"},
		{name: "no asset price baseline", hasEquipment: true, accumulatedCost: 500, wantLevel: "no_base_price"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := buildBreakEvenSummary(test.hasEquipment, test.basePrice, test.accumulatedCost)
			if got.AlertLevel != test.wantLevel {
				t.Errorf("AlertLevel = %q, want %q", got.AlertLevel, test.wantLevel)
			}
			if math.Abs(got.UsagePercent-test.wantPercent) > 1e-9 {
				t.Errorf("UsagePercent = %v, want %v", got.UsagePercent, test.wantPercent)
			}
			if got.Threshold30 != test.basePrice*0.3 || got.Threshold50 != test.basePrice*0.5 {
				t.Errorf("thresholds = %v/%v, want %v/%v", got.Threshold30, got.Threshold50, test.basePrice*0.3, test.basePrice*0.5)
			}
		})
	}
}

func TestEstimateNeedsAdminApproval(t *testing.T) {
	tests := []struct {
		name            string
		hasEquipment    bool
		basePrice       float64
		accumulatedCost float64
		estimatedCost   float64
		forceReview     bool
		want            bool
	}{
		{name: "below both thresholds", hasEquipment: true, basePrice: 1000, accumulatedCost: 200, estimatedCost: 300},
		{name: "equal to single repair threshold", hasEquipment: true, basePrice: 1000, estimatedCost: 500},
		{name: "over single repair threshold", hasEquipment: true, basePrice: 1000, estimatedCost: 500.01, want: true},
		{name: "equal to accumulated threshold", hasEquipment: true, basePrice: 1000, accumulatedCost: 500, estimatedCost: 200},
		{name: "over accumulated threshold", hasEquipment: true, basePrice: 1000, accumulatedCost: 500, estimatedCost: 200.01, want: true},
		{name: "missing equipment price baseline", hasEquipment: true, estimatedCost: 1, want: true},
		{name: "no linked equipment", estimatedCost: 1000},
		{name: "force review", hasEquipment: true, basePrice: 1000, estimatedCost: 100, forceReview: true, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := estimateNeedsAdminApproval(
				test.hasEquipment,
				test.basePrice,
				test.accumulatedCost,
				test.estimatedCost,
				test.forceReview,
			)
			if got != test.want {
				t.Errorf("estimateNeedsAdminApproval() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestBreakEvenSummaryIncludesNumbersAndHumanOnlyAdvice(t *testing.T) {
	encoded, err := json.Marshal(buildBreakEvenSummary(true, 1000, 400))
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	var response map[string]interface{}
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	for _, key := range []string{"equipment_base_price", "accumulated_repair_cost", "threshold_30", "threshold_50", "usage_percent", "risk_message", "suggested_action"} {
		if _, ok := response[key]; !ok {
			t.Errorf("summary is missing %q: %s", key, encoded)
		}
	}
	if strings.Contains(strings.ToLower(response["suggested_action"].(string)), "update automatically") {
		t.Errorf("suggestion must not imply automatic changes: %s", response["suggested_action"])
	}
}

func TestTechnicianReturnReason(t *testing.T) {
	tests := []struct {
		category string
		details  string
		want     string
		valid    bool
	}{
		{category: "ไม่ถนัด", want: "ไม่ถนัด", valid: true},
		{category: "ส่งงานผิดประเภท", details: "ควรเป็นงานระบบไฟฟ้า", want: "ส่งงานผิดประเภท: ควรเป็นงานระบบไฟฟ้า", valid: true},
		{category: "ทำไม่ได้", details: "ไม่มีเครื่องมือเฉพาะ", want: "ทำไม่ได้: ไม่มีเครื่องมือเฉพาะ", valid: true},
		{category: "", details: "เหตุผลอิสระ", valid: false},
	}

	for _, test := range tests {
		got, valid := technicianReturnReason(test.category, test.details)
		if got != test.want || valid != test.valid {
			t.Errorf("technicianReturnReason(%q, %q) = (%q, %v), want (%q, %v)", test.category, test.details, got, valid, test.want, test.valid)
		}
	}
}

func TestRepairReturnAction(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		adminNote   string
		isOwner     bool
		wantAction  string
		wantAllowed bool
	}{
		{name: "unaccepted job is rejected by owner", status: "รอซ่อม", isOwner: true, wantAction: "REJECTED_BEFORE_ACCEPT", wantAllowed: true},
		{name: "head returns delegated unaccepted job", status: "รอซ่อม", wantAction: "RETURNED_TO_ADMIN", wantAllowed: true},
		{name: "pending admin review cannot be returned", status: "รอซ่อม", adminNote: "รอแอดมินอนุมัติ", isOwner: true, wantAllowed: false},
		{name: "accepted job is returned", status: "รับงานแล้ว", isOwner: true, wantAction: "RETURNED_TO_ADMIN", wantAllowed: true},
		{name: "in progress job is returned", status: "กำลังซ่อม", isOwner: true, wantAction: "RETURNED_TO_ADMIN", wantAllowed: true},
		{name: "completed job cannot be returned", status: "เสร็จเรียบร้อย", isOwner: true, wantAllowed: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotAction, gotAllowed := repairReturnAction(test.status, test.adminNote, test.isOwner)
			if gotAction != test.wantAction || gotAllowed != test.wantAllowed {
				t.Errorf("repairReturnAction(%q, %q, %v) = (%q, %v), want (%q, %v)",
					test.status, test.adminNote, test.isOwner, gotAction, gotAllowed, test.wantAction, test.wantAllowed)
			}
		})
	}
}
