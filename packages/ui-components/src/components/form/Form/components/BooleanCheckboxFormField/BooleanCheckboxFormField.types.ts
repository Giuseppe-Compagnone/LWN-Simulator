import { FormField, FormFieldProps } from "../..";

export interface BooleanCheckboxFormFieldProps extends FormFieldProps {
  value: boolean;

  setValue(value: boolean): void;

  text: string;

  placeholder?: undefined;

  format?: undefined;

  validations?: undefined;
}

export interface BooleanCheckboxFieldOptions extends Omit<
  BooleanCheckboxFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
