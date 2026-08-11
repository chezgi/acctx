package calculator

import (
	"acctx/internal/jalali"
	"fmt"
	"sort"
)

type DeadlineRule struct {
	Event      string `json:"event"`
	TitleFA    string `json:"title_fa"`
	Operation  string `json:"operation"`
	Value      int    `json:"value"`
	SourceID   string `json:"source_id"`
}

type DeadlineResult struct {
	Event                    string `json:"event"`
	TitleFA                  string `json:"title_fa"`
	BaseDateJalali           string `json:"base_date_jalali"`
	BaselineDeadlineJalali   string `json:"baseline_deadline_jalali"`
	SourceID                 string `json:"source_id"`
	BaselineOnly             bool   `json:"baseline_only"`
	FinalVerificationNeeded  bool   `json:"final_verification_required"`
	WarningFA                string `json:"warning_fa"`
}

var deadlineRules = map[string]DeadlineRule{
	"corporate-tax-return": {
		Event: "corporate-tax-return", TitleFA: "اظهارنامه مالیات عملکرد اشخاص حقوقی", Operation: "add-months", Value: 4, SourceID: "IR-DIRECT-TAX-ARTICLE-110",
	},
	"vat-return": {
		Event: "vat-return", TitleFA: "اظهارنامه مالیات بر ارزش افزوده", Operation: "end-of-month-after", Value: 1, SourceID: "IR-VAT-LAW-ARTICLE-13",
	},
	"payroll-tax": {
		Event: "payroll-tax", TitleFA: "فهرست و پرداخت مالیات حقوق", Operation: "end-of-month-after", Value: 1, SourceID: "IR-DIRECT-TAX-ARTICLE-86",
	},
	"social-security": {
		Event: "social-security", TitleFA: "فهرست و پرداخت حق بیمه", Operation: "end-of-month-after", Value: 1, SourceID: "IR-SOCIAL-SECURITY-ARTICLE-39",
	},
	"tax-assessment-objection": {
		Event: "tax-assessment-objection", TitleFA: "اعتراض اولیه به برگ تشخیص", Operation: "add-days", Value: 30, SourceID: "IR-DIRECT-TAX-ARTICLE-238",
	},
	"tax-appeal": {
		Event: "tax-appeal", TitleFA: "اعتراض به رأی هیأت حل اختلاف", Operation: "add-days", Value: 20, SourceID: "IR-DIRECT-TAX-ARTICLE-247",
	},
}

func DeadlineEvents() []DeadlineRule {
	result := make([]DeadlineRule, 0, len(deadlineRules))
	for _, rule := range deadlineRules {
		result = append(result, rule)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Event < result[j].Event })
	return result
}

func CalculateDeadline(event, baseDate string) (DeadlineResult, error) {
	rule, ok := deadlineRules[event]
	if !ok {
		return DeadlineResult{}, fmt.Errorf("unknown deadline event %q", event)
	}
	date, err := jalali.Parse(baseDate)
	if err != nil {
		return DeadlineResult{}, err
	}
	var deadline jalali.Date
	switch rule.Operation {
	case "add-days":
		deadline, err = date.AddDays(rule.Value)
	case "add-months":
		deadline, err = date.AddMonths(rule.Value)
	case "end-of-month-after":
		deadline, err = date.EndOfMonthAfter(rule.Value)
	default:
		return DeadlineResult{}, fmt.Errorf("unsupported deadline operation")
	}
	if err != nil {
		return DeadlineResult{}, err
	}
	return DeadlineResult{
		Event:                   event,
		TitleFA:                 rule.TitleFA,
		BaseDateJalali:          date.String(),
		BaselineDeadlineJalali:  deadline.String(),
		SourceID:                rule.SourceID,
		BaselineOnly:            true,
		FinalVerificationNeeded: true,
		WarningFA:               "تعطیلات رسمی، تمدیدهای مقطعی، نحوه ابلاغ و قواعد روز کاری در این تاریخ پایه اعمال نشده‌اند.",
	}, nil
}
