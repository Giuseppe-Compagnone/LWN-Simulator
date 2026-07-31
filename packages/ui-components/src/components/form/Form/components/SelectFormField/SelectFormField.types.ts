import { JSX } from "react/jsx-runtime";
import { FormFieldProps } from "../..";

export interface SelectOption {
  value: string;
  displayed?: JSX.Element;
}

export interface SelectFormFieldProps extends FormFieldProps {
  value: string | null;
  setValue(value: string | null): void;
  options: Array<SelectOption>;
}
