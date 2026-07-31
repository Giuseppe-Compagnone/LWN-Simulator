import { ButtonProps } from "@/components/common";
import { ReactNode } from "react";

export type FormValue = string | number | boolean | Date | null | Array<string>;

export interface FormFieldProps extends Omit<FormField, "render"> {
  setValue(value: FormValue): void;
}

export interface FormField {
  name: string;
  label: string;
  value: FormValue;
  placeholder?: string;
  info?: string;

  render(props: FormFieldProps): ReactNode;
}

export interface FormProps {
  fields: FormField[];
  onSubmit(values: Record<string, FormValue>): void | Promise<void>;
  button?: ButtonProps;
}

export interface UseFormProps extends FormProps {}

export interface FormLogic {
  values: Record<string, FormValue>;
  setValue(name: string, value: FormValue): void;
}
