import { FormFieldProps } from "../..";

export interface TextFormFieldProps extends FormFieldProps {
  value: string;
  setValue(value: string): void;
  masked?: boolean;
}

export interface TextFieldOptions extends Omit<
  TextFormFieldProps,
  "render" | "setValue"
> {}
