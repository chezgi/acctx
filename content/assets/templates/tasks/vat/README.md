# VAT input semantics

- `direction`: `sale` or `purchase`.
- Sales `tax_status`: `taxable`, `exempt`, `out-of-scope`, or `unknown`.
- Purchase `tax_status`: `eligible-credit`, `noncreditable`, `exempt`, `out-of-scope`, or `unknown`.
- Amounts are integer IRR.
- Purchase VAT is included only for `eligible-credit` rows.
