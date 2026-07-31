import { FormFieldProps } from "../..";

export interface TextAreaFormFieldProps extends FormFieldProps {
  value: string;
  setValue(value: string): void;
  resize?: boolean;
  charsMax?: number;
}
