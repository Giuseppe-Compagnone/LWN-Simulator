import { FormField, FormFieldProps } from "../../Form.types";

export interface RangeFormFieldProps extends FormFieldProps {
  value: string;
  setValue(value: string): void;
  min: number;
  max: number;
  step?: number;
  formatValue?: (value: string) => string;
}

export interface RangeFieldOptions extends Omit<
  RangeFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
