---
id: acctx-evidence-bundle
version: "0.4.0"
title_fa: فهرست شواهد و بسته خروجی کنترل‌شده
---
# بسته شواهد و خروجی کنترل‌شده

## هدف
یک Task آماده را بدون تأیید، امضا یا ارسال خودکار به بسته‌ای قابل حمل، Hash‌شده و قابل Verify تبدیل کن.

## روال
1. Task، سال مالی، نوع Bundle و دامنه فایل‌ها را مشخص کن.
2. اصل فایل‌ها را در `inputs/` یا مسیرهای سال مالی حفظ کن.
3. قبل از Export، `acctx evidence index` را اجرا و فایل‌های فاقد منشأ یا ناخواسته را بررسی کن.
4. برای بسته عمومی از `acctx export task` استفاده کن.
5. برای Task حسابرسی از `acctx export audit-pack` و برای Task مالیاتی از `acctx export tax-pack` استفاده کن.
6. Bundle ساخته‌شده را با `acctx export verify` کنترل کن.
7. Digest، Bundle ID، نسخه محتوا و هشدارهای بازبینی را در مکاتبه یا Checklist ثبت کن.

## کنترل‌ها
- تمام Bundleها `draft` هستند.
- `submission_performed` همواره false است.
- Symlink و فایل غیرعادی وارد Bundle نمی‌شود.
- Export جایگزین بررسی حسابدار، مشاور مالیاتی، حسابرس یا نماینده مجاز نیست.
- قبل از ارسال، منابع و Ruleهای همان دوره دوباره Verify شوند.
