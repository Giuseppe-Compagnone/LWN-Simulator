import { FormFieldProps } from "../..";

export interface TextFormFieldProps extends FormFieldProps {
  value: string;
  setValue(value: string): void;
  masked?: boolean;
}
