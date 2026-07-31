import { JSX } from "react/jsx-runtime";
import { FormFieldProps } from "../..";

export interface CheckboxOption {
  value: string;
  displayed?: JSX.Element;
}

export interface CheckboxFormFieldProps extends FormFieldProps {
  value: Array<string> | null;
  setValue(value: Array<string> | null): void;
  options: Array<CheckboxOption>;
  placeholder?: undefined;
}

export interface CheckboxFieldOptions extends Omit<
  CheckboxFormFieldProps,
  "render" | "setValue"
> {}
