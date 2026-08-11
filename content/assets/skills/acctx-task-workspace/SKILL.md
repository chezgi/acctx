---
id: acctx-task-workspace
version: "0.2.0"
title_fa: ایجاد فضای کاری Task
---
# Task Workspace

## هدف
فایل‌های آزاد شرکت، Skill تخصصی، Template محدود و خروجی را برای یک کار مشخص به هم متصل کن.

## روال
1. Task و سال مالی را تعیین کن.
2. با `acctx task init <task> --year <year>` ساختار را بساز.
3. فایل‌های مرتبط را در `inputs/` معرفی یا کپی کن؛ اصل مدارک حفظ شود.
4. Templateهای Task را در `templates/` تکمیل کن.
5. محاسبات قطعی را در `calculations/` نگه دار.
6. متن‌ها و فرم‌های Agent را در `drafts/` بساز.
7. Checklist نهایی را تکمیل کن.

## وضعیت خروجی
تمام خروجی‌ها تا تأیید انسان `draft` هستند. `acctx` ارسال مستقیم انجام نمی‌دهد.
