# Payroll-tax input semantics

Rows are grouped by `employee_id` for an annual or selected-period reconciliation.

- `gross_taxable_irr`: ordinary salary income before the annual general exemption.
- `exempt_amount_irr`: other explicit exemptions; do not repeat the annual general exemption.
- `tax_withheld_irr`: tax actually withheld for comparison.
- Special-category and flat-rate items must be separated before using this calculator.
