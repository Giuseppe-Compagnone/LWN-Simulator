import { JSX } from "react/jsx-runtime";
import { FormFieldProps } from "../..";

export interface RadioOption {
  value: string;
  displayed?: JSX.Element;
}

export interface RadioFormFieldProps extends FormFieldProps {
  value: string | null;
  setValue(value: string | null): void;
  options: Array<RadioOption>;
  placeholder?: undefined;
}

export interface RadioFieldOptions extends Omit<
  RadioFormFieldProps,
  "render" | "setValue"
> {}
